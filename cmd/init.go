/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var dev = map[string]interface{}{
	"host":     "host",
	"port":     22,
	"username": "userName",
	"password": "password",
	"distpath": "distPath",
	"webdir":   "webDir",
	"sshtype":  "password",
}

// 判断所给路径文件/文件夹是否存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	//isnotexist来判断，是不是不存在的错误
	if os.IsNotExist(err) { //如果返回的错误类型使用os.isNotExist()判断为true，说明文件或者文件夹不存在
		return false, nil
	}
	return false, err //如果有错误了，但是不是不存在的错误，所以把这个错误原封不动的返回
}

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "初始化配置文件",
	Run: func(cmd *cobra.Command, args []string) {

		path, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		has, _ := PathExists(path + "/deploy.config.yaml")
		if has {
			fmt.Println("deploy.config.yaml 文件已存在")
			return
		}
		viper2 := viper.New()
		viper2.AddConfigPath(path) // 设置读取路径：就是在此路径下搜索配置文件。
		//viper2.AddConfigPath("$HOME/.appname")  // 多次调用以添加多个搜索路径
		viper2.SetConfigFile("deploy.config.yaml") // 设置被读取文件的全名，包括扩展名。
		viper2.SetDefault("projectname", "projectname")
		viper2.SetDefault("sshkey", "")
		viper2.SetDefault("dev", dev)
		viper2.SetDefault("prod", dev)
		viper2.ReadInConfig()
		viper2.WriteConfigAs("deploy.config.yaml")

	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
