package resourcecopyservice

import (
	"database/sql"
	"log"

	"octo/models"

	"octo/service/envservice"

	"hilo-octo-proto/go/octo"
	"github.com/pkg/errors"
)

type ResourceCopyService struct {
	resourceDao      *models.ResourceDao
	resourceUrlDao   *models.ResourceUrlDao
	versionDao       *models.VersionDao
	adminResourceDao *models.AdminResourceDao
	tagDao           *models.TagDao
}

type CopySelectedFileOptions struct {
	AppId                int
	SourceVersionId      int
	DestinationVersionId int
	DestinationAppId     int
	Filenames            []string
	DryRun               bool
}

var envService = envservice.NewEnvService()

func NewResourceCopyService() *ResourceCopyService {
	return &ResourceCopyService{
		resourceDao:      models.NewResourceDao(),
		resourceUrlDao:   models.NewResourceUrlDao(),
		versionDao:       &models.VersionDao{},
		adminResourceDao: models.NewAdminResourceDao(),
		tagDao:           &models.TagDao{},
	}
}

func (s *ResourceCopyService) CopySelectedFile(o CopySelectedFileOptions, envCheck bool) (map[string]string, error) {
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

	//  实际复制时
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
	}
	//// 实际复制时
	//if !o.DryRun {
	//
	//	srcVersion, err := s.versionDao.Get(o.AppId, o.SourceVersionId)
	//	if err != nil {
	//		return nil, errors.Wrap(err, "failed to get src version")
	//	}
	//
	//	if srcVersion.EnvId.Int64 > 0 {
	//		if err := s.versionDao.AddVersionWithEnvId(destAppId, o.DestinationVersionId, int(srcVersion.EnvId.Int64)); err != nil {
	//			return nil, errors.Wrap(err, "failed to add version")
	//		}
	//	} else {
	//		if err := s.versionDao.AddVersion(destAppId, o.DestinationVersionId); err != nil {
	//			return nil, errors.Wrap(err, "failed to add version")
	//		}
	//	}
	//}
	tx, err := models.StartTransaction()
	if err != nil {
		return nil, err
	}
	res, err := s.copySelectedFile(o, tx)
	log.Printf("[INFO] CopySelectedFile result: %+v\n", res)
	return res, models.FinishTransaction(tx, err)
}

func (s *ResourceCopyService) copySelectedFile(o CopySelectedFileOptions, tx *sql.Tx) (map[string]string, error) {

	var newRevision int
	var destAppId int
	if o.DestinationAppId > 0 {
		destAppId = o.DestinationAppId
	} else {
		destAppId = o.AppId
	}

	// 实际复制时
	if !o.DryRun {
		var err error

		newRevision, err = s.versionDao.IncrementMaxRevision(destAppId, o.DestinationVersionId, tx)

		if err != nil {
			return nil, err
		}
	}
	res := map[string]string{}
	for _, name := range o.Filenames {
		file, err := s.resourceDao.GetByNameForUpdate(o.AppId, o.SourceVersionId, name, tx)
		if err != nil {
			return nil, err
		}
		if (file == models.Resource{}) {
			return nil, errors.Errorf("resource not found: %v", name)
		}

		err = s.copyAppOneFile(file, destAppId, o.DestinationVersionId, o.DryRun, newRevision, res, tx)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

func (s *ResourceCopyService) copyAppOneFile(file models.Resource, targetAppId int, targetVersion int, dryRun bool, newRevision int, res map[string]string, tx *sql.Tx) error {
	tfile, err := s.resourceDao.GetByNameForUpdate(targetAppId, targetVersion, file.Filename, tx)
	if err != nil {
		return err
	}
	if file.State == int(octo.Data_DELETE) {
		if (tfile == models.Resource{}) {
			res[file.Filename] = "already_deleted"
			return nil
		}
	}

	fileUrl, err := s.resourceUrlDao.GetUrlByObjectNameAndRevisionIdLatest(file.AppId, file.VersionId, file.ObjectName.String, file.RevisionId)
	if err != nil {
		return err
	}
	if (fileUrl == models.ResourceUrl{}) && file.State != int(octo.Data_DELETE) {
		return errors.Errorf("missing fileUrl: %v", file.Filename)
	}

	tfileUrl, err := s.resourceUrlDao.GetUrlByObjectNameLatest(file.AppId, targetVersion, file.ObjectName.String)
	if err != nil {
		return err
	}
	if fileUrl.Md5 == tfileUrl.Md5 && file.Tag == tfile.Tag && file.State == tfile.State {
		res[file.Filename] = "already_exists"
		return nil
	}

	// 实际复制时
	if !dryRun {
		f := file
		f.AppId = targetAppId
		f.VersionId = targetVersion
		f.RevisionId = newRevision

		targetIdFile, err := s.adminResourceDao.GetByIdFromTx(targetAppId, targetVersion, file.Id, tx)
		if err != nil {
			return err
		}

		if tfile.Id > 0 || (targetIdFile == models.Resource{}) {
			// 复制目标已经存在同名资产，或复制目标相同ID没有使用的情况下相同ID使用
			if tfile.Id > 0 {
				f.Id = tfile.Id
			}
			err = s.adminResourceDao.Replace(f, tx)
		} else {
			err = s.adminResourceDao.InsertWithId(f, tx)
		}
		if err != nil {
			return err
		}

		if (fileUrl != models.ResourceUrl{}) {
			u := fileUrl
			u.AppId = targetAppId
			u.VersionId = targetVersion
			u.RevisionId = newRevision
			err = s.resourceUrlDao.AddUrl(u, tx)
			if err != nil {
				return err
			}
		}
	}
	res[file.Filename] = "copied"
	return nil
}
