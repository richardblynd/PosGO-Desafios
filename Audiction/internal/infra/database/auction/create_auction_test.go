package auction

import (
	"context"
	"fullcycle-auction_go/internal/entity/auction_entity"
	"os"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestAuctionAutoClose(t *testing.T) {
	mongoURL := os.Getenv("MONGODB_URL")
	if mongoURL == "" {
		mongoURL = "mongodb://admin:admin@localhost:27017/auctions_test?authSource=admin"
	}

	os.Setenv("AUCTION_INTERVAL", "3s")

	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURL))
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(ctx)

	if err := client.Ping(ctx, nil); err != nil {
		t.Skipf("MongoDB not available, skipping integration test: %v", err)
	}

	database := client.Database("auctions_test")
	database.Collection("auctions").Drop(ctx)
	defer database.Collection("auctions").Drop(ctx)

	repo := NewAuctionRepository(database)

	auctionEntity, ierr := auction_entity.CreateAuction(
		"Test Product",
		"Electronics",
		"Test description for auction auto-close",
		auction_entity.New,
	)
	if ierr != nil {
		t.Fatalf("Failed to create auction entity: %s", ierr.Error())
	}

	if ierr := repo.CreateAuction(ctx, auctionEntity); ierr != nil {
		t.Fatalf("Failed to insert auction: %s", ierr.Error())
	}

	// Verify auction is initially Active
	var auctionMongo AuctionEntityMongo
	filter := bson.M{"_id": auctionEntity.Id}
	if err := repo.Collection.FindOne(ctx, filter).Decode(&auctionMongo); err != nil {
		t.Fatalf("Failed to find auction: %v", err)
	}
	if auctionMongo.Status != auction_entity.Active {
		t.Fatalf("Expected auction status to be Active (0), got %d", auctionMongo.Status)
	}

	// Wait for the auction interval to expire (3s + 1s buffer)
	t.Log("Waiting for auction to auto-close...")
	time.Sleep(4 * time.Second)

	// Verify the auction status changed to Completed
	if err := repo.Collection.FindOne(ctx, filter).Decode(&auctionMongo); err != nil {
		t.Fatalf("Failed to find auction after interval: %v", err)
	}
	if auctionMongo.Status != auction_entity.Completed {
		t.Fatalf("Expected auction status to be Completed (1), got %d", auctionMongo.Status)
	}

	t.Log("Auction was automatically closed successfully!")
}
