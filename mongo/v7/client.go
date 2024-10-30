package v7

import (
	"fmt"
	"github.com/spf13/viper"
	"github.com/topfreegames/extensions/v9/mongo/interfaces"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Client struct {
	Config  *viper.Viper
	MongoDB interfaces.MongoDB
}

func NewClient(prefix string, config *viper.Viper, mongoOrNil ...interfaces.MongoDB) (*Client, error) {
	client := &Client{
		Config: config,
	}
	var mongoDB interfaces.MongoDB
	if len(mongoOrNil) > 0 {
		mongoDB = mongoOrNil[0]
	}
	err := client.Connect(prefix, mongoDB)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func makeKey(prefix, sufix string) string {
	return fmt.Sprintf("%s.%s", prefix, sufix)
}

// Connect connects to mongo db and saves on Client
func (c *Client) Connect(prefix string, db interfaces.MongoDB) error {
	if db != nil {
		c.MongoDB = db
		return nil
	}
	url := c.Config.GetString(makeKey(prefix, "url"))
	user := c.Config.GetString(makeKey(prefix, "user"))
	pass := c.Config.GetString(makeKey(prefix, "pass"))
	database := c.Config.GetString(makeKey(prefix, "db"))
	timeout := c.Config.GetDuration(makeKey(prefix, "connectionTimeout"))

	var client *mongo.Client
	var err error

	clientOpts := []*options.ClientOptions{
		options.Client().ApplyURI(url),
		options.Client().SetAuth(options.Credential{
			Username: user,
			Password: pass,
		}),
	}

	if timeout > 0 {
		clientOpts = append(clientOpts, options.Client().SetConnectTimeout(timeout))
	}

	client, err = mongo.Connect(nil, clientOpts...)
	if err != nil {
		return err
	}

	// TODO: allow configuring read and write concern
	mongoDB := client.Database(database)
	c.MongoDB = newMongo(client, mongoDB)

	return nil
}

// Close closes the session and the connection with db
func (c *Client) Close() {
	c.MongoDB.Close()
}
