package model

import gcpstorage "google.golang.org/api/storage/v1"

type FirebaseStorage struct {
	Objects    *gcpstorage.ObjectsService
	BucketName string
}
