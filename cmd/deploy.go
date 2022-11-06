/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"github.com/pkg/sftp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
)

func uploadFile(sftpClient *sftp.Client, localFilePath string, remotePath string) error {
	//打开本地文件流
	srcFile, err := os.Open(localFilePath)
	if err != nil {
		return err
	}
	//关闭文件流
	defer srcFile.Close()
	// 上传到远端服务器的文件名,与本地路径末尾相同
	var remoteFileName = path.Base(localFilePath)
	// 打开远程文件,如果不存在就创建一个
	dstFile, err := sftpClient.Create(path.Join(remotePath, remoteFileName))
	if err != nil {
		return err

	}
	//关闭远程文件
	defer dstFile.Close()
	//读取本地文件,写入到远程文件中(这里没有分快穿,自己写的话可以改一下,防止内存溢出)
	ff, err := io.ReadAll(srcFile)
	if err != nil {
		return err

	}
	dstFile.Write(ff)
	return nil
}
func uploadDirectory(sftpClient *sftp.Client, localPath string, remotePath string) error {
	//打开本地文件夹流
	localFiles, err := os.ReadDir(localPath)
	if err != nil {
		return err
	}
	//先创建最外层文件夹
	sftpClient.Mkdir(remotePath)
	//遍历文件夹内容
	for _, backupDir := range localFiles {
		localFilePath := path.Join(localPath, backupDir.Name())
		remoteFilePath := path.Join(remotePath, backupDir.Name())
		//判断是否是文件,是文件直接上传.是文件夹,先远程创建文件夹,再递归复制内部文件
		if backupDir.IsDir() {
			sftpClient.Mkdir(remoteFilePath)
			err := uploadDirectory(sftpClient, localFilePath, remoteFilePath)
			if err != nil {
				return err
			}
		} else {
			uploadFile(sftpClient, path.Join(localPath, backupDir.Name()), remotePath)
		}
	}

	return nil
}
func Upload(sftpClient *sftp.Client, localPath string, remotePath string) error {
	//获取路径的属性
	s, err := os.Stat(localPath)
	if err != nil {
		fmt.Printf("%c[%d;%d;%dm%s%c[0m \n", 0x1B, 31, 40, 1, "文件路径不存在", 0x1B)
		return err
	}
	//判断是否是文件夹
	fmt.Printf("%c[%d;%d;%dm%s%c[0m \n", 0x1B, 32, 40, 1, "2.文件上传中...", 0x1B)
	if s.IsDir() {
		err := uploadDirectory(sftpClient, localPath, remotePath)
		if err != nil {
			return err
		}
	} else {
		err := uploadFile(sftpClient, localPath, remotePath)
		if err != nil {
			return err
		}
	}
	return nil
}

// func publicKeyAuthFunc(kPath string) ssh.AuthMethod {
// 	keyPath, err := homedir.Expand(kPath)
// 	if err != nil {
// 		log.Fatal("find key's home dir failed", err)
// 	}
// 	key, err := io.ReadFile(keyPath)
// 	if err != nil {
// 		log.Fatal("ssh key file read failed", err)
// 	}
// 	// Create the Signer for this private key.
// 	signer, err := ssh.ParsePrivateKey(key)
// 	if err != nil {
// 		log.Fatal("ssh key signer failed", err)
// 	}
// 	return ssh.PublicKeys(signer)
// }

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "启动部署",
	Long:  `通过--mode指定部署环境`,
	Run: func(cmd *cobra.Command, args []string) {
		mode, _ := cmd.Flags().GetString("mode")
		pwdpath, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		config := viper.New()
		config.AddConfigPath(pwdpath)         //设置读取的文件路径
		config.SetConfigName("deploy.config") //设置读取的文件名
		config.SetConfigType("yaml")          //设置文件的类型
		//尝试进行配置读取
		if err := config.ReadInConfig(); err != nil {
			panic(err)
		}
		projectname := config.Get("projectname")
		// sshkey := config.Get("sshkey")
		sshtype := config.Get(mode + ".sshtype")
		distpath := config.Get(mode + ".distpath")
		sshHost := config.Get(mode + ".host")
		sshPort := config.Get(mode + ".port")
		username := config.Get(mode + ".username")
		password := config.Get(mode + ".password")
		webdir := config.Get(mode + ".webdir")

		sshConfig := &ssh.ClientConfig{
			Timeout:         time.Second * 20, //ssh 连接time out 时间一秒钟, 如果ssh验证错误 会在一秒内返回
			User:            fmt.Sprintf("%v", username),
			HostKeyCallback: ssh.InsecureIgnoreHostKey(), //这个可以， 但是不够安全
			//HostKeyCallback: hostKeyCallBackFunc(h.Host),
		}
		if sshtype == "password" {
			sshConfig.Auth = []ssh.AuthMethod{ssh.Password(fmt.Sprintf("%v", password))}
		} else {
			// sshConfig.Auth = []ssh.AuthMethod{publicKeyAuthFunc(sshkey)}
		}
		// dial 获取ssh client
		addr := fmt.Sprintf("%s:%d", sshHost, sshPort)
		sshClient, err := ssh.Dial("tcp", addr, sshConfig)
		if err != nil {
			fmt.Printf("%c[%d;%d;%dm%s%c[0m \n", 0x1B, 31, 40, 1, "ssh连接失败", 0x1B)
		}

		defer sshClient.Close()

		// 创建ssh-session
		session, err := sshClient.NewSession()
		if err != nil {
			str := fmt.Sprintf("创建ssh session 失败:%V", err)
			fmt.Printf(" %c[%d;%d;%dm%s%c[0m \n", 0x1B, 31, 40, 1, str, 0x1B)

		}
		fmt.Printf("%c[%d;%d;%dm%s%c[0m \n", 0x1B, 32, 40, 1, "1.ssh链接成功", 0x1B)
		defer session.Close()
		// 执行远程命令
		// combo, err := session.CombinedOutput(fmt.Sprintf("cd %v;rm -rf %v; mkdir %v", webdir, projectname, projectname))
		// if err != nil {
		// 	str := fmt.Sprintf("远程执行cd ;rm -fr ;mkdir;失败:%v;%v", err, combo)
		// 	fmt.Printf(" %c[%d;%d;%dm%s%c[0m \n", 0x1B, 31, 40, 1, str, 0x1B)
		// 	return
		// }
		// str := fmt.Sprintf("2.当前目录%v,%v已经删除", webdir, projectname)
		// fmt.Printf("%c[%d;%d;%dm%s%c[0m \n", 0x1B, 32, 40, 1, str, 0x1B)

		sftpClient, err := sftp.NewClient(sshClient)
		if err != nil {
			str := fmt.Sprintf("sftp链接失败%v", err)
			fmt.Printf("%c[%d;%d;%dm%s%c[0m \n ", 0x1B, 31, 40, 1, str, 0x1B)
		}
		defer sftpClient.Close()

		var localDir = fmt.Sprintf("%v", distpath)                    //本地文件目录
		var remoteFilePath = fmt.Sprintf("%v%v", webdir, projectname) //远程文件目录

		err = Upload(sftpClient, localDir, remoteFilePath)
		if err != nil {
			str := fmt.Sprintf("上传失败%v", err)
			fmt.Printf("%c[%d;%d;%dm%s%c[0m \n ", 0x1B, 31, 40, 1, str, 0x1B)
			return
		}
		// ========= done
		fmt.Printf("%c[%d;%d;%dm%s%c[0m \n", 0x1B, 32, 40, 1, "3.上传成功", 0x1B)
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)
	// Here you will define your flags and configuration settings.
	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// deployCmd.PersistentFlags().String("mode", "dev", "A help for foo")
	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	deployCmd.Flags().StringP("mode", "m", "dev", "请输入部署环境")
}
