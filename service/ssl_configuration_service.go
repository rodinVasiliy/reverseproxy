package service

import (
	config "reverseproxy/config/mongo_config"
	ssl "reverseproxy/model/ssl_config"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SSLConfigurationService struct {
	deps *config.MongoDeps
}

func NewSSLConfigurationService(deps *config.MongoDeps) *SSLConfigurationService {
	return &SSLConfigurationService{deps: deps}
}

func (sslS *SSLConfigurationService) FindByName(name string) (*ssl.SSLConfiguration, error) {
	return findByName[ssl.SSLConfiguration](sslS.deps, sslS.deps.Config.Database, SSL_COLLECTION, name)
}

func (sslS *SSLConfigurationService) FindById(id primitive.ObjectID) (*ssl.SSLConfiguration, error) {
	return findById[ssl.SSLConfiguration](sslS.deps, sslS.deps.Config.Database, SSL_COLLECTION, id)
}

func (sslS *SSLConfigurationService) Add(sslConfig *ssl.SSLConfiguration) (primitive.ObjectID, error) {
	return add(sslS.deps, sslS.deps.Config.Database, SSL_COLLECTION, sslConfig)
}
