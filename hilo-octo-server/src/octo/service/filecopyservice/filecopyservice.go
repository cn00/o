package filecopyservice

import (
	"database/sql"
	"log"
	"octo/models"
	"octo/service/envservice"
	"octo/utils"

	"hilo-octo-proto/go/octo"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type FileCopyService struct {
	fileDao      *models.FileDao
	fileUrlDao   *models.FileUrlDao
	versionDao   *models.VersionDao
	adminFileDao *models.AdminFileDao
	tagDao       *models.TagDao
}

type CopySelectedFileOptions struct {
	AppId                int
	SourceVersionId      int
	DestinationAppId     int
	DestinationVersionId int
	Filenames            []string
	DryRun               bool
}

type result struct {
	message string
	fileID  int
}

var envService = envservice.NewEnvService()

func NewFileCopyService() *FileCopyService {
	return &FileCopyService{
		fileDao:      models.NewFileDao(),
		fileUrlDao:   models.NewFileUrlDao(),
		versionDao:   &models.VersionDao{},
		adminFileDao: models.NewAdminFileDao(),
		tagDao:       &models.TagDao{},
	}
}

func (s *FileCopyService) CopySelectedFile(o CopySelectedFileOptions, envCheck bool) (map[string]string, error) {
	log.Printf("[INFO] CopySelectedFile options: %+v\n", o)

	// 没有选择一个副本的情况
	if len(o.Filenames) == 0 {
		return nil, errors.New("Please select copy source")
	}

	var destAppId int
	if o.DestinationAppId > 0 {
		destAppId = o.DestinationAppId
	} else {
		destAppId = o.AppId
	}

	err := envService.CheckSameEnvironment(o.AppId, o.SourceVersionId, destAppId, o.DestinationVersionId, envCheck)
	if err != nil {
		return nil, err
	}

	// 德莱兰false(=实际复制)的情况
	if !o.DryRun {
		srcVersion, err := s.versionDao.Get(o.AppId, o.SourceVersionId)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get src version")
		}

		if srcVersion.EnvId.Valid && int(srcVersion.EnvId.Int64) > 0 {

			dstEnv, err := envService.GetSameEnv(o.AppId, destAppId, srcVersion.VersionId)
			if err != nil {
				return nil, errors.Wrap(err, "get same env")
			}

			// dstのEnvId的情况下EnvIdもCopy、dst如果没有srcのenvId利用
			// Copy前面的version如果不存在EnvId也创建
			if dstEnv.EnvId > 0 {
				if err := s.versionDao.AddVersionWithEnvId(destAppId, o.DestinationVersionId, dstEnv.EnvId); err != nil {
					return nil, errors.Wrap(err, "failed to add version")
				}
			} else {
				if err := s.versionDao.AddVersionWithEnvId(destAppId, o.DestinationVersionId, int(srcVersion.EnvId.Int64)); err != nil {
					return nil, errors.Wrap(err, "failed to add version")
				}
			}
		} else {
			// Copy前面的version不存在时Version作成（EnvId无）
			if err := s.versionDao.AddVersion(destAppId, o.DestinationVersionId); err != nil {
				return nil, errors.Wrap(err, "failed to add version")
			}
		}
		//if srcVersion.EnvId.Int64 > 0 {
		//	if err := s.versionDao.AddVersionWithEnvId(destAppId, o.DestinationVersionId, int(srcVersion.EnvId.Int64)); err != nil {
		//		return nil, errors.Wrap(err, "failed to add version")
		//	}
		//} else {
		//	if err := s.versionDao.AddVersion(destAppId, o.DestinationVersionId); err != nil {
		//		return nil, errors.Wrap(err, "failed to add version")
		//	}
		//}
	}
	tx, err := models.StartTransaction()
	if err != nil {
		return nil, err
	}
	res, err := s.copySelectedFile(o, tx)
	log.Printf("[INFO] CopySelectedFile result: %+v\n", res)
	return res, models.FinishTransaction(tx, err)
}

func (s *FileCopyService) copySelectedFile(o CopySelectedFileOptions, tx *sql.Tx) (map[string]string, error) {

	var destAppId int
	if o.DestinationAppId > 0 {
		destAppId = o.DestinationAppId
	} else {
		destAppId = o.AppId
	}

	var newRevision int
	// 德莱兰false(=实际复制)的情况
	if !o.DryRun {
		var err error
		newRevision, err = s.versionDao.IncrementMaxRevision(destAppId, o.DestinationVersionId, tx)

		if err != nil {
			return nil, err
		}
	}

	copyResults := map[string]result{}
	for _, name := range o.Filenames {
		file, err := s.fileDao.GetByNameForUpdate(o.AppId, o.SourceVersionId, name, tx)
		if err != nil {
			return nil, err
		}
		if (file == models.File{}) {
			return nil, errors.Errorf("file not found: %v", name)
		}
		err = s.copyAppOneFile(file, destAppId, o.DestinationVersionId, o.DryRun, newRevision, copyResults, tx)

		if err != nil {
			return nil, err
		}
	}

	res := map[string]string{}
	for k, v := range copyResults {
		res[k] = v.message
	}
	return res, nil
}

func (s *FileCopyService) copyAppOneFile(file models.File, targetAppId int, targetVersion int, dryRun bool, newRevision int, res map[string]result, tx *sql.Tx) error {
	if _, ok := res[file.Filename]; ok {
		if gin.IsDebugging() {
			log.Println("[DEBUG]already visited:", file.Filename)
		}
		return nil
	}

	tfile, err := s.fileDao.GetByNameForUpdate(targetAppId, targetVersion, file.Filename, tx)
	if err != nil {
		return err
	}
	if (tfile == models.File{}) {
		if file.State == int(octo.Data_DELETE) {
			res[file.Filename] = result{"already_deleted", -1}
			return nil
		}
	}

	fileUrl, err := s.fileUrlDao.GetUrlByObjectNameAndRevisionIdLatest(file.AppId, file.VersionId, file.ObjectName.String, file.RevisionId)
	if err != nil {
		return err
	}
	if (fileUrl == models.FileUrl{}) && file.State != int(octo.Data_DELETE) {
		return errors.Errorf("missing fileUrl: %v", file.Filename)
	}

	tfileUrl, err := s.fileUrlDao.GetUrlByObjectNameLatest(targetAppId, targetVersion, file.ObjectName.String)
	if err != nil {
		return err
	}
	if fileUrl.Crc == tfileUrl.Crc && file.Tag == tfile.Tag && file.State == tfile.State {
		res[file.Filename] = result{"already_exists", tfile.Id}
		return nil
	}

	if file.Dependency.Valid {
		ids, err := utils.SplitDependencies(file.Dependency.String)
		if err != nil {
			return err
		}
		newDeps := make([]int, 0, len(ids)) // copy有时ID因为有改变的可能性，所以重新制作复制后的依赖信息
		for _, id := range ids {
			dfile, err := s.adminFileDao.GetById(file.AppId, file.VersionId, id)
			if err != nil {
				return err
			}
			if (dfile == models.File{}) || dfile.State == int(octo.Data_DELETE) {
				return errors.Errorf("missing dependency: file=%v, dependent file=%v", file.Filename, dfile.Filename)
			}
			err = s.copyAppOneFile(dfile, targetAppId, targetVersion, dryRun, newRevision, res, tx)
			if err != nil {
				return err
			}
			newDeps = append(newDeps, res[dfile.Filename].fileID)
		}
		file.Dependency.String = utils.JoinDependencies(newDeps)
	}

	// 实际复制时
	var copyID int
	if !dryRun {
		f := file
		f.AppId = targetAppId
		f.VersionId = targetVersion
		f.RevisionId = newRevision

		targetIdFile, err := s.adminFileDao.GetByIdFromTx(targetAppId, targetVersion, file.Id, tx)
		if err != nil {
			return err
		}

		if tfile.Id > 0 || (targetIdFile == models.File{}) {
			// 复制目标已经存在同名资产，或复制目标相同ID没有使用的情况下相同ID使用
			if tfile.Id > 0 {
				f.Id = tfile.Id
			}
			copyID = f.Id
			err = s.adminFileDao.Replace(f, tx)
		} else {
			copyID, err = s.adminFileDao.InsertWithId(f, tx)
		}
		if err != nil {
			return err
		}

		if (fileUrl != models.FileUrl{}) {
			u := fileUrl
			u.AppId = targetAppId
			u.VersionId = targetVersion
			u.RevisionId = newRevision
			err := s.fileUrlDao.AddUrl(u, tx)
			if err != nil {
				return err
			}
		}
	}
	res[file.Filename] = result{"copied", copyID}
	return nil
}
