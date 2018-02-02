package cmd

import (
	"log"
	"strconv"

	"github.com/lbryio/reflector.go/db"
	"github.com/lbryio/reflector.go/peer"
	"github.com/lbryio/reflector.go/store"

	"github.com/spf13/cobra"
)

func init() {
	var peerCmd = &cobra.Command{
		Use:   "peer",
		Short: "Run peer server",
		Run:   peerCmd,
	}
	RootCmd.AddCommand(peerCmd)
}

func peerCmd(cmd *cobra.Command, args []string) {
	db := new(db.SQL)
	err := db.Connect(GlobalConfig.DBConn)
	checkErr(err)

	s3 := store.NewS3BlobStore(GlobalConfig.AwsID, GlobalConfig.AwsSecret, GlobalConfig.BucketRegion, GlobalConfig.BucketName)
	combo := store.NewDBBackedS3Store(s3, db)
	log.Fatal(peer.NewServer(combo).ListenAndServe("localhost:" + strconv.Itoa(peer.DefaultPort)))
}
