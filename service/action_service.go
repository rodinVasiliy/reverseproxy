package service

import (
	"fmt"
	config "reverseproxy/config/mongo_config"
	actiondoc "reverseproxy/model/action"

	"go.mongodb.org/mongo-driver/bson"
)

type ActionService struct {
	deps *config.MongoDeps
}

func NewActionService(deps *config.MongoDeps) *ActionService {
	return &ActionService{deps: deps}
}

func LoadActionsToDB(as *ActionService) error {
	logToDBActionDoc := actiondoc.ActionDoc{
		Name: LOG_TO_DB_ACTION_NAME,
	}
	blockRequestActionDoc := actiondoc.ActionDoc{
		Name: BLOCK_REQUEST_ACTION_NAME,
	}

	db := as.deps.Config.Database
	collection := as.deps.Client.Database(db).Collection(ACTIONS_COLLECTION)
	ctx, cancel := as.deps.Ctx()
	defer cancel()

	_, err := collection.InsertOne(ctx, logToDBActionDoc)
	if err != nil {
		return err
	}
	_, err = collection.InsertOne(ctx, blockRequestActionDoc)
	if err != nil {
		return err
	}
	return nil
}

func FindAllActions(as *ActionService) ([]actiondoc.ActionDoc, error) {
	db := as.deps.Config.Database
	collection := as.deps.Client.Database(db).Collection(ACTIONS_COLLECTION)
	ctx, cancel := as.deps.Ctx()
	defer cancel()

	// Находим все документы
	cursor, err := collection.Find(ctx, bson.M{}) // bson.M{} — пустой фильтр = "всё"
	if err != nil {
		fmt.Errorf("failed to find actions %s", err)
	}
	defer cursor.Close(ctx)

	var actions []actiondoc.ActionDoc

	// Декодируем все документы в срез структур
	if err := cursor.All(ctx, &actions); err != nil {
		fmt.Errorf("failed to decode actions %s", err)
	}

	return actions, nil
}
