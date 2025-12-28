package ssl

import (
	"path/filepath"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SSLConfiguration struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	Name         string             `bson:"name"`
	CertFileName string             `bson:"cert"`
	KeyFileName  string             `bson:"key"`
}

var SSL_FILES_PATH = filepath.Join("etc", "nginx", "ssl")

func (sslConfig *SSLConfiguration) GetID() primitive.ObjectID {
	return sslConfig.ID
}
