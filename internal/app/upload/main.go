package upload

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"bitbucket.org/airenas/listgo/internal/pkg/msgsender"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"bitbucket.org/airenas/listgo/internal/pkg/saver"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "uploadService",
	Short: "LiST Upload Audio File Service",
	Long:  `HTTP server to listen and upload audio files for transcription`,
	Run:   run,
}

func init() {
	cmdapp.InitApplication(rootCmd)
	rootCmd.PersistentFlags().Int32P("port", "", 8000, "Default service port")
	viper.BindPFlag("port", rootCmd.PersistentFlags().Lookup("port"))
	viper.SetDefault("port", 8080)
	viper.SetDefault("fileStorage.path", "/data/audio.in/")
}

//Execute starts the server
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) {
	log.Println("Starting uploadService")
	log.Println("Init File Storage for " + viper.GetString("fileStorage.path"))
	fileSaver := saver.NewLocalFileSaver(viper.GetString("fileStorage.path"))
	msgSender := new(msgsender.MachineMessageSender)
	StartWebServer(&ServiceData{fileSaver, msgSender, strconv.Itoa(viper.GetInt("port"))})
}
