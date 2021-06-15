package service

import (
	"database/sql"
	"hilo-octo-proto/go/octo"
	"log"
	"octo/service/envservice"
	"time"

	"octo/models"
	"octo/utils"

	"github.com/pkg/errors"
)

var envService = envservice.NewEnvService()

type SyncService struct {
}

func (service *SyncService) DiffSyncLatest(dstAppId int, dstVersionId int, srcAppId int, srcVersionId int) error {
	//check revision
	maxRevision, err := versionDao.GetMaxRevision(srcAppId, srcVersionId)
	if err != nil {
		return errors.Wrap(err, "failed to get maxrevision")
	}
	log.Printf("[INFO] SyncService: DiffSyncLatest %+v", struct {
		Source      interface{}
		Destination interface{}
	}{
		Source: struct {
			AppId       int
			VersionId   int
			MaxRevision int
		}{
			AppId:       srcAppId,
			VersionId:   srcVersionId,
			MaxRevision: maxRevision,
		},
		Destination: struct {
			AppId     int
			VersionId int
		}{
			AppId:     dstAppId,
			VersionId: dstVersionId,
		},
	})
	return service.DiffSync(dstAppId, dstVersionId, srcAppId, srcVersionId, maxRevision)
}

func (s *SyncService) DiffSync(dstAppId int, dstVersionId int, srcAppId int, srcVersionId int, revisionId int) error {

	tx, err := models.StartTransaction()
	if err != nil {
		return err
	}
	return models.FinishTransaction(tx, s.diffSync(dstAppId, dstVersionId, srcAppId, srcVersionId, revisionId, tx))
}

func (s *SyncService) diffSync(dstAppId int, dstVersionId int, srcAppId int, srcVersionId int, revisionId int, tx *sql.Tx) error {
	srcFileList, dstCurrentFileMap, srcResourceList, dstCurrentResourceMap, maxRevision, err := s.diffSyncStart(dstAppId, dstVersionId, srcAppId, srcVersionId, revisionId)
	if err != nil {
		return err
	}

	var tags []string

	for _, srcFile := range srcFileList {
		dstFile, err := s.diffSyncOneFile(dstAppId, dstVersionId, srcFile, maxRevision, revisionId, dstCurrentFileMap, tx)
		if err != nil {
			return err
		}
		if (dstFile != models.File{}) {
			tags = utils.MergeTags(tags, utils.SplitTags(dstFile.Tag.String))
		}
	}

	for _, srcResource := range srcResourceList {
		destResource, err := s.diffSyncOneResource(dstAppId, dstVersionId, srcResource, maxRevision, revisionId, dstCurrentResourceMap, tx)
		if err != nil {
			return err
		}
		if (destResource != models.Resource{}) {
			tags = utils.MergeTags(tags, utils.SplitTags(destResource.Tag.String))
		}
	}

	if err := tagService.ensureTags(dstAppId, tags, tx); err != nil {
		return err
	}

	return s.diffSyncEnd(dstAppId, dstVersionId, revisionId, tx)
}

func (*SyncService) diffSyncOneFile(dstAppId int, dstVersionId int, srcFile models.File, maxRevision int,
	revisionId int, currentFileMap map[string]models.File, tx *sql.Tx) (models.File, error) {
	// sync无需
	if srcFile.RevisionId <= maxRevision {
		return models.File{}, nil
	}

	dstFile, ok := currentFileMap[srcFile.Filename]
	if !ok && srcFile.State == int(octo.Data_DELETE) {
		// sync开元delete立着旗子sync如果不先存在的话，为了防止事故sync不做
		return models.File{}, nil
	}

	//revision靠，靠URL取得
	srcUrl, err := fileUrlDao.GetUrlByObjectNameAndRevisionIdLatest(srcFile.AppId, srcFile.VersionId,
		srcFile.ObjectName.String, revisionId)
	if err != nil {
		return models.File{}, errors.Wrap(err, "failed to get url")
	}
	if (srcUrl == models.FileUrl{}) {
		log.Println("[WARN] SyncService: srcFile url not found")
	}
	log.Printf("[INFO] SyncService: copy ab:　%+v\n", struct {
		Source      interface{}
		Destination interface{}
	}{
		Source: struct {
			AppId             int
			VersionId         int
			RevisionId        int
			ObjectName        string
			Filename          string
			FileRevisionId    int
			FileUrlRevisionId int
			BuildNumber       string
			UploadVersionId   int
		}{
			AppId:             srcFile.AppId,
			VersionId:         srcFile.VersionId,
			RevisionId:        revisionId,
			ObjectName:        srcFile.ObjectName.String,
			Filename:          srcFile.Filename,
			FileRevisionId:    srcFile.RevisionId,
			FileUrlRevisionId: srcUrl.RevisionId,
			BuildNumber:       srcFile.BuildNumber.String,
			UploadVersionId:   int(srcFile.UploadVersionId.Int64),
		},
		Destination: struct {
			AppId       int
			VersionId   int
			MaxRevision int
		}{
			AppId:       dstAppId,
			VersionId:   dstVersionId,
			MaxRevision: maxRevision,
		},
	})

	//file更新，否则创建
	if !ok {
		//insert
		dstFile = models.File{
			Item: models.Item{
				Id:              srcFile.Id,
				AppId:           dstAppId,
				VersionId:       dstVersionId,
				RevisionId:      srcFile.RevisionId,
				Filename:        srcFile.Filename,
				ObjectName:      srcFile.ObjectName,
				Size:            srcFile.Size,
				Generation:      srcFile.Generation,
				Md5:             srcUrl.Md5,
				Tag:             srcFile.Tag,
				Priority:        srcFile.Priority,
				State:           srcFile.State,
				BuildNumber:     srcFile.BuildNumber,
				UploadVersionId: srcFile.UploadVersionId,
			},
			Crc:        srcUrl.Crc,
			Assets:     srcFile.Assets,
			Dependency: srcFile.Dependency,
		}
		err := fileDao.Insert(dstFile, tx)
		if err != nil {
			return models.File{}, errors.Wrap(err, "failed to insert srcFile")
		}
	} else {
		dstFile.Sync(srcFile)
		_, err := fileDao.Update(dstFile, tx)
		if err != nil {
			return models.File{}, errors.Wrap(err, "failed to update srcFile")
		}
	}
	url := models.FileUrl{
		AppId:       dstAppId,
		VersionId:   dstVersionId,
		RevisionId:  srcUrl.RevisionId,
		ObjectName:  srcUrl.ObjectName,
		Crc:         srcUrl.Crc,
		Md5:         srcUrl.Md5,
		Url:         srcUrl.Url,
		UpdDatetime: time.Now(),
	}
	err = fileUrlDao.AddUrl(url, tx)
	if err != nil {
		return models.File{}, errors.Wrap(err, "failed to insert fileUrl")
	}
	return dstFile, nil
}

func (*SyncService) diffSyncOneResource(dstAppId int, dstVersionId int, srcResource models.Resource, maxRevision int,
	revisionId int, dstCurrentResourceMap map[string]models.Resource, tx *sql.Tx) (models.Resource, error) {
	// sync无需
	if srcResource.RevisionId <= maxRevision {
		return models.Resource{}, nil
	}

	dstFile, ok := dstCurrentResourceMap[srcResource.Filename]
	if !ok && srcResource.State == int(octo.Data_DELETE) {
		// sync开元delete立着旗子sync如果不先存在的话，为了防止事故sync不做
		return models.Resource{}, nil
	}

	//revision靠，靠URL取得
	srcUrl, err := resourceUrlDao.GetUrlByObjectNameAndRevisionIdLatest(srcResource.AppId, srcResource.VersionId, srcResource.ObjectName.String, revisionId)
	if err != nil {
		return models.Resource{}, errors.Wrap(err, "failed to get url")
	}
	if (srcUrl == models.ResourceUrl{}) {
		log.Println("[WARN] SyncService: resource url not found")
	}
	log.Printf("[INFO] SyncService: copy r:　%+v\n", struct {
		Source      interface{}
		Destination interface{}
	}{
		Source: struct {
			AppId             int
			VersionId         int
			RevisionId        int
			ObjectName        string
			Filename          string
			FileRevisionId    int
			FileUrlRevisionId int
			BuildNumber       string
			UploadVersionId   int
		}{
			AppId:             srcResource.AppId,
			VersionId:         srcResource.VersionId,
			RevisionId:        revisionId,
			ObjectName:        srcResource.ObjectName.String,
			Filename:          srcResource.Filename,
			FileRevisionId:    srcResource.RevisionId,
			FileUrlRevisionId: srcUrl.RevisionId,
			BuildNumber:       srcResource.BuildNumber.String,
			UploadVersionId:   int(srcResource.UploadVersionId.Int64),
		},
		Destination: struct {
			AppId       int
			VersionId   int
			MaxRevision int
		}{
			AppId:       dstAppId,
			VersionId:   dstVersionId,
			MaxRevision: maxRevision,
		},
	})

	if !ok {
		//insert
		dstFile = models.Resource{
			Item: models.Item{
				Id:              srcResource.Id,
				AppId:           dstAppId,
				VersionId:       dstVersionId,
				RevisionId:      srcResource.RevisionId,
				Filename:        srcResource.Filename,
				ObjectName:      srcResource.ObjectName,
				Size:            srcResource.Size,
				Generation:      srcResource.Generation,
				Md5:             srcUrl.Md5,
				Tag:             srcResource.Tag,
				Priority:        srcResource.Priority,
				State:           srcResource.State,
				UpdDatetime:     srcResource.UpdDatetime,
				BuildNumber:     srcResource.BuildNumber,
				UploadVersionId: srcResource.UploadVersionId,
			},
		}
		err := resourceDao.Insert(dstFile, tx)
		if err != nil {
			return models.Resource{}, errors.Wrap(err, "failed to insert resource")
		}
	} else {
		dstFile.Sync(srcResource.Item)
		_, err := resourceDao.Update(dstFile, tx)
		if err != nil {
			return models.Resource{}, errors.Wrap(err, "failed to update resource")
		}
	}
	dstResrcUrl := models.ResourceUrl{
		AppId:       dstAppId,
		VersionId:   dstVersionId,
		RevisionId:  srcUrl.RevisionId,
		ObjectName:  srcUrl.ObjectName,
		Md5:         srcUrl.Md5,
		Url:         srcUrl.Url,
		UpdDatetime: srcUrl.UpdDatetime,
	}
	err = resourceUrlDao.AddUrl(dstResrcUrl, tx)
	if err != nil {
		return models.Resource{}, errors.Wrap(err, "failed to insert resourceUrl")
	}
	return dstFile, nil
}

func (*SyncService) diffSyncStart(dstAppId int, dstVersionId int, srcAppId int, srcVersionId int, revisionId int) ([]models.File, map[string]models.File, []models.Resource, map[string]models.Resource, int, error) {

	//check revision
	maxRevision, err := versionDao.GetMaxRevision(dstAppId, dstVersionId)
	if err != nil {
		return nil, nil, nil, nil, 0, errors.Wrap(err, "failed to get maxrevision")
	}

	if maxRevision > revisionId {
		return nil, nil, nil, nil, 0, errors.New("invalid src revisionId (regression)")
	}

	srcFileList, err := fileDao.GetDiffList(srcAppId, srcVersionId, 0, revisionId)
	if err != nil {
		return nil, nil, nil, nil, 0, errors.Wrap(err, "failed to get list of file")
	}

	srcResourceList, err := resourceDao.GetDiffList(srcAppId, srcVersionId, 0, revisionId)
	if err != nil {
		return nil, nil, nil, nil, 0, errors.Wrap(err, "failed to get list of resource")
	}

	srcMaxRevision, err := versionDao.GetMaxRevision(srcAppId, srcVersionId)
	if err != nil {
		return nil, nil, nil, nil, 0, errors.Wrap(err, "failed to get source maxrevision")
	}

	if srcMaxRevision < revisionId {
		return nil, nil, nil, nil, 0, errors.New("invalid source revisionId (not found)")
	}

	// Copy前面的Env为了确认Copy前面的Version正在获取
	srcVersion, err := versionDao.Get(srcAppId, srcVersionId)
	if err != nil {
		return nil, nil, nil, nil, 0, errors.Wrap(err, "failed to get version")
	}

	if srcVersion.EnvId.Valid && int(srcVersion.EnvId.Int64) > 0 {

		dstEnv, err := envService.GetSameEnv(srcAppId, dstAppId, srcVersion.VersionId)
		if err != nil {
			return nil, nil, nil, nil, 0, errors.Wrap(err, "get same env")
		}

		// dstのEnvId的情况下EnvIdもCopy、dst如果没有srcのenvId利用
		// Copy前面的version如果不存在EnvId也创建
		if dstEnv.EnvId > 0 {
			if err := versionDao.AddVersionWithEnvId(dstAppId, dstVersionId, dstEnv.EnvId); err != nil {
				return nil, nil, nil, nil, 0, errors.Wrap(err, "failed to add version")
			}
		} else {
			if err := versionDao.AddVersionWithEnvId(dstAppId, dstVersionId, int(srcVersion.EnvId.Int64)); err != nil {
				return nil, nil, nil, nil, 0, errors.Wrap(err, "failed to add version")
			}
		}
	} else {
		// Copy前面的version不存在时Version作成（EnvId无）
		if err := versionDao.AddVersion(dstAppId, dstVersionId); err != nil {
			return nil, nil, nil, nil, 0, errors.Wrap(err, "failed to add version")
		}
	}

	dstCurrentFileList, err := fileDao.GetList(dstAppId, dstVersionId, 0)
	if err != nil {
		return nil, nil, nil, nil, 0, errors.Wrap(err, "failed to get list of current file")
	}
	dstCurrentFileMap := map[string]models.File{}
	for _, file := range dstCurrentFileList {
		dstCurrentFileMap[file.Filename] = file
	}

	dstCurrentResourceList, err := resourceDao.GetList(dstAppId, dstVersionId, 0)
	if err != nil {
		return nil, nil, nil, nil, 0, errors.Wrap(err, "failed to get list of resource")
	}
	dstCurrentResourceMap := map[string]models.Resource{}
	for _, file := range dstCurrentResourceList {
		dstCurrentResourceMap[file.Filename] = file
	}
	return srcFileList, dstCurrentFileMap, srcResourceList, dstCurrentResourceMap, maxRevision, nil
}

func (*SyncService) diffSyncEnd(dstAppId int, dstVersionId int, revisionId int, tx *sql.Tx) error {
	err := versionDao.UpdateMaxRevision(dstAppId, dstVersionId, revisionId, tx)
	if err != nil {
		return errors.Wrap(err, "failed to update MaxRevision")
	}
	return nil
}
