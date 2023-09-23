package mongodb

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/sr-codefreak/user-group/myerrors"
	"github.com/sr-codefreak/user-group/utils/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
)

var log = logger.GetLogger()

type mongoclient struct {
	sync.RWMutex
	c                      *mongo.Client
	isConnected            bool
	isDisconnecting        bool
	shouldDisconnect       bool
	isMonitoringConnection bool
}

func (mc *mongoclient) waitForDisconnecting() bool {
	for mc._getIsDisconnecting() {
		time.Sleep(1 * time.Second)
	}
	return true
}

func (mc *mongoclient) _getIsDisconnecting() bool {
	return mc.isDisconnecting
}

func (mc *mongoclient) getIsConnected() bool {
	mc.RLock()
	defer mc.RUnlock()
	return mc.waitForDisconnecting() && mc.isConnected
}

func (mc *mongoclient) setIsConnected(conn bool) {
	if mc.getIsConnected() != conn {
		mc.Lock()
		defer mc.Unlock()
		mc.isConnected = conn
		if conn {
			mc.isDisconnecting = false
		}
	}
}

func (mc *mongoclient) getIsMonitoringConnection() bool {
	mc.Lock()
	defer mc.Unlock()
	return mc.waitForDisconnecting() && mc.isMonitoringConnection
}

func (mc *mongoclient) setIsMonitoringConnection(monitoring bool) {
	if mc.getIsMonitoringConnection() != monitoring {
		mc.Lock()
		defer mc.Unlock()
		mc.isMonitoringConnection = monitoring
	}
}

func (mc *mongoclient) getShouldDisconnect() bool {
	mc.RLock()
	defer mc.RUnlock()
	return mc.waitForDisconnecting() && mc.shouldDisconnect
}

func (mc *mongoclient) setShouldDisconnect(disconn bool) {
	if mc.getShouldDisconnect() != disconn {
		mc.Lock()
		defer mc.Unlock()
		mc.shouldDisconnect = disconn
	}
}

func (mc *mongoclient) getClient() *mongo.Client {
	mc.RLock()
	defer mc.RUnlock()
	return mc.c
}

func (mc *mongoclient) setClient(c *mongo.Client) {
	mc.Lock()
	defer mc.Unlock()
	mc.c = c
}

func (mc *mongoclient) disconnect() error {
	mc.Lock()
	defer mc.Unlock()
	mc.isDisconnecting = true
	var err error
	if mc.c != nil {
		err = mc.c.Disconnect(context.Background())
		if err != nil {
			if err == mongo.ErrClientDisconnected {
				log.Warnf("Mongo client already disconnected")
			} else {
				log.Errorf("error in disconnect: %s", err)
			}
		}
		mc.c = nil
	}
	mc.isConnected = false
	mc.isDisconnecting = false
	return err
}

func (mc *mongoclient) markForDisconnect() {
	mc.setShouldDisconnect(true)
	log.Warnf("mongo connection marked for disconnection")
}

var client = mongoclient{
	c:                      nil,
	isConnected:            false,
	isDisconnecting:        false,
	shouldDisconnect:       false,
	isMonitoringConnection: false,
}

// Connect connects and sets a connection.
// More info on mongo connection string
// https://docs.mongodb.com/manual/reference/connection-string/
func Connect(connectionURI string, dbSigChan chan struct{}) error {
	client.setShouldDisconnect(false)
	notified := false
	if !client.getIsMonitoringConnection() {
		go func() {
			for !client.getShouldDisconnect() {
				isConnectedOld := client.getIsConnected()
				if !client.getIsConnected() {
					connect(connectionURI)
				}
				if client.getClient() != nil {
					err := ping()
					if err != nil {
						log.Warnf("ping error: %s", err)
						_ = client.disconnect()
						if isConnectedOld != client.getIsConnected() {
							//log only when connection status changes
							log.Errorf("cannot create mongo session: %s\n", err)
						}
					} else {
						client.setIsConnected(true)
						if isConnectedOld != client.getIsConnected() {
							//log only when connection status changes
							cs, err := connstring.ParseAndValidate(connectionURI)
							if err == nil {
								var hostsStr string
								for i, h := range cs.Hosts {
									hostsStr += h
									if i != (len(cs.Hosts) - 1) {
										hostsStr += ","
									}
								}
								log.Info("Connected to mongo at " + hostsStr)
								if !notified && dbSigChan != nil {
									dbSigChan <- struct{}{}
									notified = true
								}
							}
						}
					}
				} else {
					client.setIsConnected(false)
				}
				time.Sleep(5 * time.Second)
			}
			_ = client.disconnect()
			client.setIsMonitoringConnection(false)
		}()
		client.setIsMonitoringConnection(true)
	} else {
		for !client.getIsConnected() {
			time.Sleep(1 * time.Second)
		}
		go func() {
			if !notified && dbSigChan != nil {
				dbSigChan <- struct{}{}
				notified = true
			}
		}()
	}
	return nil
}

func ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return client.c.Ping(ctx, readpref.Primary())
}

func connect(connectionURI string) error {
	if !client.getIsConnected() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		c, err := mongo.Connect(ctx, options.Client().ApplyURI(connectionURI))
		if err != nil {
			client.setIsConnected(false)
			return err
		}
		client.setClient(c)
	}
	return nil
}

// Disconnect mongo client connection.
// Returns myerrors.ErrNoMongoConnection as error when no mongo connection
func Disconnect() error {
	if !client.getIsConnected() {
		return myerrors.ErrNoMongoConnection
	}
	client.markForDisconnect()
	return nil
}

// GetClient returns the connected mongo client
// Returns myerrors.ErrNoMongoConnection as error when no mongo connection
func GetClient() (*mongo.Client, error) {
	if !client.getIsConnected() {
		return nil, myerrors.ErrNoMongoConnection
	}
	return client.getClient(), nil
}

type JSONSerializer interface {
	SerializeToJSON(w http.ResponseWriter) error
}

type JSONBodyDeserializer interface {
	DeserializeFromJSONInBody(r *http.Request) JSONBodyDeserializer
}

type CollectionNamer interface {
	CollectionName() string
}

type DatabaseNamer interface {
	DatabaseName() string
}

type collectionDatabaseNamer interface {
	CollectionNamer
	DatabaseNamer
}

type UserGroup struct {
	DatabaseNamer
}

func (a UserGroup) DatabaseName() string {
	return "userGroup"
}

// FindOne finds one entry in a collection based on mongo query
// Returns myerrors.myerrors.ErrNoMongoConnection as error when no mongo connection
func FindOne(m collectionDatabaseNamer, i interface{}, query bson.D, opts ...*options.FindOneOptions) (bool, error) {
	if !client.getIsConnected() {
		return false, myerrors.ErrNoMongoConnection
	}
	c := client.getClient().Database(m.DatabaseName()).Collection(m.CollectionName())
	result := c.FindOne(nil, query, opts...)
	err := result.Decode(i)
	if err != nil && err != mongo.ErrNoDocuments {
		return false, err
	}
	if err == mongo.ErrNoDocuments {
		return false, nil
	}
	return true, nil
}
func FindOneWithModel(m collectionDatabaseNamer, i interface{}, d bson.D) (interface{}, bool, error) {
	if !client.getIsConnected() {
		return i, false, myerrors.ErrNoMongoConnection
	}
	c := client.getClient().Database(m.DatabaseName()).Collection(m.CollectionName())
	result := c.FindOne(nil, d)
	err := result.Decode(i)
	if err != nil && err != mongo.ErrNoDocuments {
		return i, false, err
	}
	if err == mongo.ErrNoDocuments {
		return i, false, nil
	}
	return i, true, nil
}

// UpdateOne finds one entry in a collection based on mongo query and updates it
// Returns myerrors.ErrNoMongoConnection as error when no mongo connection
func UpdateOne(m collectionDatabaseNamer, filter bson.D, update bson.D, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	if !client.getIsConnected() {
		return nil, myerrors.ErrNoMongoConnection
	}
	c := client.getClient().Database(m.DatabaseName()).Collection(m.CollectionName())
	result, err := c.UpdateOne(nil, filter, bson.D{{"$set", update}}, opts...)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// UpdateMany finds all entry in a collection based on mongo query and updates it
// Returns myerrors.ErrNoMongoConnection as error when no mongo connection
func UpdateMany(m collectionDatabaseNamer, filter bson.D, update bson.D, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	if !client.getIsConnected() {
		return nil, myerrors.ErrNoMongoConnection
	}
	c := client.getClient().Database(m.DatabaseName()).Collection(m.CollectionName())
	result, err := c.UpdateMany(nil, filter, bson.D{{"$set", update}}, opts...)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func AddToArray(m collectionDatabaseNamer, filter bson.D, update bson.D, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	if !client.getIsConnected() {
		return nil, myerrors.ErrNoMongoConnection
	}
	c := client.getClient().Database(m.DatabaseName()).Collection(m.CollectionName())
	result, err := c.UpdateOne(nil, filter, bson.D{{"$addToSet", update}}, opts...)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// DeleteOne delete one entry in a collection based on mongo query
// Returns myerrors.ErrNoMongoConnection as error when no mongo connection
func DeleteOne(m collectionDatabaseNamer, d bson.D) error {
	if !client.getIsConnected() {
		return myerrors.ErrNoMongoConnection
	}
	c := client.getClient().Database(m.DatabaseName()).Collection(m.CollectionName())
	dr, err := c.DeleteOne(nil, d)
	if err != nil {
		return err
	}
	if dr.DeletedCount == 1 {
		return nil
	}
	return errors.New("did not delete exactly one document")
}

// DeleteMany delete many entries in a collection based on mongo query
// Returns myerrors.ErrNoMongoConnection as error when no mongo connection
func DeleteMany(m collectionDatabaseNamer, d bson.D) error {
	if !client.getIsConnected() {
		return myerrors.ErrNoMongoConnection
	}
	c := client.getClient().Database(m.DatabaseName()).Collection(m.CollectionName())
	count, err := c.DeleteMany(nil, d)
	if err != nil {
		return err
	}
	logger.NewLogger().Println("deleted the documents", count.DeletedCount)
	return nil
}

func DeleteAll(m collectionDatabaseNamer) error {
	if !client.getIsConnected() {
		return myerrors.ErrNoMongoConnection
	}
	c := client.getClient().Database(m.DatabaseName()).Collection(m.CollectionName())
	err := c.Drop(nil)
	if err != nil {
		return err
	}
	return nil
}

// InsertOne inserts one entry in a collection based on mongo query.
// Returns myerrors.ErrNoMongoConnection as error when no mongo connection
func InsertOne(m collectionDatabaseNamer, i interface{}) (interface{}, error) {
	if !client.getIsConnected() {
		return nil, myerrors.ErrNoMongoConnection
	}
	c := client.getClient().Database(m.DatabaseName()).Collection(m.CollectionName())
	ir, err := c.InsertOne(nil, i)
	if err != nil {
		return nil, err
	}
	return ir.InsertedID, nil
}

// InsertMany inserts many entry in a collection based on mongo query.
// Returns myerrors.ErrNoMongoConnection as error when no mongo connection
func InsertMany(m collectionDatabaseNamer, docs []interface{}) ([]interface{}, error) {
	if !client.getIsConnected() {
		return nil, myerrors.ErrNoMongoConnection
	}
	c := client.getClient().Database(m.DatabaseName()).Collection(m.CollectionName())
	ir, err := c.InsertMany(nil, docs)
	if err != nil {
		return nil, err
	}
	return ir.InsertedIDs, nil
}

func CountDocuments(m collectionDatabaseNamer, filter bson.D) (int64, error) {
	if !client.getIsConnected() {
		return 0, myerrors.ErrNoMongoConnection
	}
	c := client.getClient().Database(m.DatabaseName()).Collection(m.CollectionName())
	result, err := c.CountDocuments(nil, filter)
	if err != nil && err != mongo.ErrNoDocuments {
		return 0, err
	}
	if err == mongo.ErrNoDocuments {
		return 0, nil
	}
	return result, nil
}

func Find(m collectionDatabaseNamer, filter interface{}, opts ...*options.FindOptions) (*mongo.Cursor, error) {
	if !client.getIsConnected() {
		return nil, myerrors.ErrNoMongoConnection
	}
	c := client.getClient().Database(m.DatabaseName()).Collection(m.CollectionName())
	cursor, err := c.Find(nil, filter, opts...)
	if err != nil && err != mongo.ErrNoDocuments {
		return nil, err
	}
	return cursor, nil
}

func Aggregate(m collectionDatabaseNamer, d mongo.Pipeline) (*mongo.Cursor, error) {
	if !client.isConnected {
		return nil, myerrors.ErrNoMongoConnection
	}
	c := client.c.Database(m.DatabaseName()).Collection(m.CollectionName())
	cursor, err := c.Aggregate(nil, d)
	if err != nil && err != mongo.ErrNoDocuments {
		return nil, err
	}
	return cursor, nil
}

// // AuthenticateWithJWT authenticates user with jwt token
// func AuthenticateWithJWT(id interface{}, r *http.Request) (*User, bool, error) {
// 	_id := id.(string)
// 	objID, err := primitive.ObjectIDFromHex(_id)
// 	if err != nil {
// 		return nil, false, err
// 	}

// 	user := &User{}
// 	userExists, err := FindOne(userModel, user, bson.D{{Key: "_id", Value: objID}})
// 	if err != nil || !userExists {
// 		return nil, false, err
// 	}
// 	return user, true, nil
// }

func GetBsonDForStruct(structData interface{}) (bson.D, error) {
	pByte, err := bson.Marshal(structData)
	if err != nil {
		return nil, err
	}

	var update bson.D
	err = bson.Unmarshal(pByte, &update)
	if err != nil {
		return nil, err
	}

	return update, nil
}

func GetBsonAForArray(arrayData interface{}) (bson.A, error) {
	in := struct{ Data interface{} }{Data: arrayData}
	inData, err := bson.Marshal(in)
	if err != nil {
		return nil, err
	}
	var out struct{ Data bson.A }
	if err := bson.Unmarshal(inData, &out); err != nil {
		return nil, err
	}
	return out.Data, nil
}
func Distinct(m collectionDatabaseNamer, field string, filter interface{}) ([]interface{}, error) {
	if !client.getIsConnected() {
		return nil, myerrors.ErrNoMongoConnection
	}
	c := client.getClient().Database(m.DatabaseName()).Collection(m.CollectionName())
	data, err := c.Distinct(nil, field, filter)
	if err != nil && err != mongo.ErrNoDocuments {
		return nil, err
	}
	return data, nil
}

// Update with unset key removes the key and value on passing
// update-obj with bson.D{{"$unset", bson.D{{"<key>", ""}}}}
func UpdateWithUnsetKey(m collectionDatabaseNamer, filter bson.D, update bson.D, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	if !client.getIsConnected() {
		return nil, myerrors.ErrNoMongoConnection
	}
	c := client.getClient().Database(m.DatabaseName()).Collection(m.CollectionName())
	result, err := c.UpdateOne(nil, filter, update, opts...)
	if err != nil {
		return nil, err
	}
	return result, nil
}
