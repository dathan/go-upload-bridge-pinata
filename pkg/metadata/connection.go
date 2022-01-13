package metadata

import (
	"os"

	goshard "github.com/dathan/go-shard"
	"github.com/sirupsen/logrus"
)

//implements goshard.ShardConfig
type ShardConfig struct{}

//helper function to use a OS ENV KEY or value
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func (s ShardConfig) GetShardLookup() goshard.ShardLookup {

	var cp []*goshard.ConnectionParams
	var i uint = 1
	cp = append(cp, &goshard.ConnectionParams{
		Host:     getEnv("DB_HOST", "127.0.0.1"),
		Dbname:   getEnv("DB_NAME", "foreverawards"),
		User:     getEnv("DB_USER", "foreveraward"),
		Password: getEnv("DB_PASS", "yoyoma"),
		ShardId:  i,
	})

	logrus.Infof("DBInfo: %+v", *cp[0])

	rsl := goshard.NewShardLookup(cp)
	return rsl

}

//set up all the hosts
func (sconf ShardConfig) NewShardConnection(entity_id uint64) (error, *goshard.ShardConnection) {

	rsl := sconf.GetShardLookup()

	err, c := rsl.Lookup(entity_id)
	if err != nil {
		return err, nil
	}

	err, rc := goshard.NewConnection(c)
	sc := &goshard.ShardConnection{*rc, c.ShardId}

	return err, sc

}

//Get a new shard connection by id
func (sconf ShardConfig) NewShardConnectionByShardId(shard_id uint) (error, *goshard.ShardConnection) {

	rsl := sconf.GetShardLookup()
	cs := rsl.GetAll()
	shard_id = shard_id - 1 // 0 based
	c := cs[shard_id]
	err, rc := goshard.NewConnection(c)
	sc := &goshard.ShardConnection{*rc, c.ShardId}

	return err, sc

}
