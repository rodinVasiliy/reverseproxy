package utils

import (
	"fmt"
	"net"
	"net/http"
	"os/exec"
	config "reverseproxy/internal/infrastructure/config/mongo_config"

	"go.mongodb.org/mongo-driver/bson"
)

func GetIpFromRequest(r *http.Request) net.IP {
	xrip := r.Header.Get("X-Real-IP")
	return net.ParseIP(xrip)
}

func DropAllCollections(deps *config.MongoDeps) error {
	ctx, cancel := deps.Ctx()
	defer cancel()

	db := deps.Client.Database(deps.Config.Database)

	coll, err := db.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		return err
	}

	for _, c := range coll {
		if err := db.Collection(c).Drop(ctx); err != nil {
			return fmt.Errorf("failed to drop %s collection %w", c, err)
		}
	}
	return nil
}

// удаляет файлы nginx для вебапов(если использовать удаление вебапа, то файл удалится сам, но при чистке всей базы эти файлы надо удалять вручную)
func DropOldWebappFiles() {
	exec.Command("bash", "-c", "rm -f /etc/nginx/sites-available/webapp-*").Run()
	exec.Command("bash", "-c", "rm -f /etc/nginx/sites-enabled/webapp-*").Run()
}
