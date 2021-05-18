package service

import (
	"log"

	"octo/models"
)

var fileUrlDao = models.NewFileUrlDao()
var resourceUrlDao = models.NewResourceUrlDao()

type MaintenanceService struct {
}

func (*MaintenanceService) MakeDiffSql(appId int, versionId int, tappId int, tversionId int) {

	log.Println("start makediffsql")
	//get list
	fileList, _ := adminFileDao.GetList(appId, versionId, "", "", "", nil, nil, nil, 0, "", "", true)
	currentFileMap := map[string]models.File{}
	for _, object := range fileList {
		file := object.(models.File)
		if file.ObjectName.String == "" {
			continue
		}
		currentFileMap[file.Filename] = file
	}

	maxRevision, _ := versionDao.GetMaxRevision(appId, versionId)
	log.Println(maxRevision)
	tmaxRevision, _ := versionDao.GetMaxRevision(tappId, tversionId)
	log.Println(tmaxRevision)
	log.Printf("SELECT maxRevision+1 INTO m FROM version where app_id=%d and version_id=%d;", tmaxRevision, tversionId)
	//get tlist
	tfileList, _ := adminFileDao.GetList(tappId, tversionId, "", "", "", nil, nil, nil, 0, "", "", true)
	for _, object := range tfileList {
		tfile := object.(models.File)
		if tfile.ObjectName.String == "" {
			continue
		}
		file, ok := currentFileMap[tfile.Filename]
		if ok {
			if file.RevisionId < tmaxRevision && tfile.RevisionId < tmaxRevision {
				if file.State != tfile.State {
					log.Printf("UPDATE files SET state=%d,revision_id=@m WHERE app_id=%d AND version_id=%d AND id=%d and filename='%s';", file.State, tfile.AppId, tfile.VersionId, tfile.Id, tfile.Filename)

				}
			}
		}

	}
	log.Printf("UPDATE version SET max_revision=@m where app_id=%d and version_id=%d;", tmaxRevision, tversionId)
}
