package registryservice

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/izuc/asset-registry/pkg/registry"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Service struct {
	db *mongo.Database
}

func NewService(mongoDB *mongo.Database) *Service {
	return &Service{db: mongoDB}
}

func (s *Service) SaveAsset(ctx context.Context, network string, asset *registry.Asset) error {
	_, err := s.db.Collection(network).InsertOne(ctx, asset)
	return errors.Wrap(err, "failed to insert assets into mongo collection")
}

func (s *Service) LoadAssets(ctx context.Context, network string, IDs ...string) (assets []*registry.Asset, err error) {
	var cur *mongo.Cursor
	assets = make([]*registry.Asset, 0)
	if len(IDs) == 0 {
		cur, err = s.db.Collection(network).Find(ctx, bson.M{})
		if err != nil {
			return
		}

		defer cur.Close(context.TODO())

		for cur.Next(context.TODO()) {
			var asset *registry.Asset
			err = cur.Decode(&asset)
			if err != nil {
				return
			}
			assets = append(assets, asset)
		}
		if err = cur.Err(); err != nil {
			return
		}
		return
	}

	for _, ID := range IDs {
		// Query One
		result := s.db.Collection(network).FindOne(ctx, bson.M{"ID": ID})
		var asset *registry.Asset
		err = result.Decode(&asset)
		if err != nil {
			return
		}
		assets = append(assets, asset)
	}

	return
}

func (s *Service) LoadAsset(ctx context.Context, network string, ID string) (asset *registry.Asset, err error) {
	// Query One
	result := s.db.Collection(network).FindOne(ctx, bson.M{"ID": ID})
	err = result.Decode(&asset)
	if err != nil {
		return
	}

	return
}

func (s *Service) DeleteAssetByID(ctx context.Context, network string, ID string) (err error) {
	// Delete all
	_, err = s.db.Collection(network).DeleteMany(ctx, bson.M{"ID": ID})
	return
}

func (s *Service) DeleteAssetByName(ctx context.Context, network string, name string) (err error) {
	// Delete all
	_, err = s.db.Collection(network).DeleteMany(ctx, bson.M{"name": name})
	return
}
