package cmd

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	//"encoding/hex"
	"github.com/CosmicBDry/helm-push/cipher"
	"github.com/CosmicBDry/helm-push/tools"
	"github.com/spf13/cobra"
)

var (
	AppVersion     bool
	CipherPassword string
	ChartDir       string
	ReleaseTag     string
	Content        string
	User           string
)

var (
	keyText       = "abcdefgehjhijkmlkjjwwoewrtyuisdg"
	bindFileBytes = make([]byte, 1024)
	objMap        = make(map[string]any)
)

var RootCmd = &cobra.Command{
	Use:   "helm-push",
	Short: "Helm Plugin, Charts push to remote helm repository",
	Long:  "Helm Plugin, Charts push to remote helm repository",

	RunE: func(cmd *cobra.Command, args []string) error {
		if AppVersion {
			//1.打印工具的版本信息
			fmt.Printf("Describe: %s\nGoVersion: %s\nAuthor: %s\nAppVersion: %s\n", "helm-push plugin(require helmV3.0+)",
				"go1.20.14", "zhaojiehe", "v1.0.1")
			os.Exit(0)

		} else if len(CipherPassword) > 0 {
			//2.命令行密码加密工具
			cipherByte, err := cipher.Encrypt(CipherPassword, keyText)
			if err != nil {
				return err
			}
			fmt.Printf("%X\n", cipherByte)
			os.Exit(0)
		} else if len(ChartDir) > 0 { //3.以下为helm push的整个过程
			//1).chart包版本号生成----------------------------------------------------------------------
			ChartDir = strings.TrimLeft(ChartDir, "./")
			ChartDir = strings.TrimRight(ChartDir, "/")
			DefaultTag := "v" + time.Now().Format("200601021504") + "+" + ReleaseTag
			ChartFile := ChartDir + "/" + "Chart.yaml"

			//2).读取Chart.yaml文件，将内容解析到map实例中
			bindFileBytes, err := tools.ReadFile(ChartFile, bindFileBytes, 0644) //bindFileBytes, err := tools.OsReadFile(ChartFile)

			if err != nil {
				return err
			}

			tools.YamlUnmarshalMap(bindFileBytes, objMap)
			if v, ok := objMap["annotations"].(map[string]any)["helm.push/plugin-enable"].(bool); ok {
				if v != true {
					fmt.Println("helm.push/plugin is not enable")
					os.Exit(-1)
				}
			} else if v, ok := objMap["annotations"].(map[string]any)["helm.push/plugin-enable"].(string); ok {
				V, _ := strconv.ParseBool(v)
				if V != true {
					fmt.Println("helm.push/plugin is not enable")
					os.Exit(-1)
				}

			}
			//3).修改chart包版本信息并更新到Chart.yaml------------------------------------------------------
			objMap["version"] = DefaultTag
			objMap["appVersion"] = "V" + ReleaseTag
			bindFileBytes, err = tools.YamlMarshalMap(objMap)
			if err != nil {
				return err
			}
			err = tools.WriteFile(ChartFile, bindFileBytes, 0644)
			if err != nil {
				return err
			}
			//4).提交信息更新到Release.info文件------------------------------------------------------------
			projectName := objMap["annotations"].(map[string]any)["helm.sh/project"].(string)
			modeName := objMap["name"]
			commitContent := fmt.Sprintf("提交时间: %s\n项目名称: %s\n所属模块: %s\nChart当前版本: %s\n提交人: %s\n更新内容:\n%s",
				time.Now().Format("2006-01-02TZ15:04:05"),
				projectName, modeName, DefaultTag, User,
				tools.LineIndentBuidler(Content, "    "))
			err = tools.WriteFileCreate(ChartDir+"/"+"Release.info", commitContent, 0644)
			if err != nil {
				return fmt.Errorf("tools.WriteFileCreate(Release.info) Error: %s", err.Error())
			}
			//5).执行helm本地打包--------------------------------------------------------------------------
			localCmd := fmt.Sprintf("helm package %s", ChartDir)
			output, err := tools.LocalCommand(localCmd)
			fmt.Println(output)
			if err != nil {
				return fmt.Errorf("helm package error: %s", err.Error())
			}

			//6).基于环境变量获取ssh连接的user、password、host------------------------------------------------
			ciperbytes, err := cipher.HexDecrypt(os.Getenv("HELM_REPO_ATUH_TOKEN"))
			if err != nil {
				return err
			}
			plainPassword, err := cipher.Decrypt(ciperbytes, keyText)
			if err != nil {
				return fmt.Errorf("HELM_REPO_ATUH_TOKEN 错误: %s", err.Error())
			}
			username := os.Getenv("HELM_REPO_AUTH_USER")
			password := plainPassword
			remoteAddr := os.Getenv("HELM_REPO_HOST") + ":" + "22"
			//7).创建ssh的client远程连接---------------------------------------------------------------------
			client, err := tools.CreateSshClient(username, password, remoteAddr)
			if err != nil {
				return err
			}
			defer client.Close()

			//8).本地、远程等相关文件路径变量定义----------------------------------------------------------------
			ssh_path := objMap["annotations"].(map[string]any)["helm.ssh/path"].(string)
			http_path := objMap["annotations"].(map[string]any)["helm.http/path"].(string)
			http_port := 0
			if port, ok := objMap["annotations"].(map[string]any)["helm.http/port"].(int); ok {
				http_port = port
			} else if port, ok := objMap["annotations"].(map[string]any)["helm.http/port"].(string); ok {
				http_port, _ = strconv.Atoi(port)
			}
			localFilePath := ChartDir + "-" + DefaultTag + ".tgz"
			remoteFilePath := ssh_path + "/" + localFilePath
			//9).发送chart压缩包至远程服务----------------------------------------------------------------------
			err = tools.SendFileToRemote(localFilePath, remoteFilePath, client)
			if err != nil {
				return fmt.Errorf("helm-push failure: %s", err.Error())
			}
			//10).执行远程系统命令，对helm仓库进行index索引创建-----------------------------------------------------
			remoteCmd := fmt.Sprintf("helm repo index %s --url=http://%s:%d%s",
				ssh_path, os.Getenv("HELM_REPO_HOST"), http_port, http_path)

			err = tools.RemoteCommand(remoteCmd, client)
			if err != nil {
				return err
			}
			//11).执行成功后，打印成功结果至控制台---------------------------------------------------------------
			fmt.Printf("HELM-PUSH <%s> Completed!!!, The Helm-Chart is as follows:\n\n>>>[\"http://%s:%d%s/%s\"]\n",
				localFilePath,
				os.Getenv("HELM_REPO_HOST"),
				http_port,
				http_path,
				localFilePath)

			os.Exit(0)
		}
		return errors.New("invalid arguments or incorrect useage")
	},
}

func Execute() {
	RootCmd.Execute()
}

// 选项参数引入
func init() {
	RootCmd.PersistentFlags().BoolVarP(&AppVersion, "version", "v", false, "版本信息")
	RootCmd.PersistentFlags().StringVarP(&CipherPassword, "encrypt", "e", "", "密码加密并打印输出")
	RootCmd.PersistentFlags().StringVarP(&ChartDir, "package", "p", "", "推送至helm仓库的Chart包名称(一个Chart包对应一个目录)")
	RootCmd.PersistentFlags().StringVarP(&ReleaseTag, "tag", "t", "1.0.0", "Chart包发布版本")
	RootCmd.PersistentFlags().StringVarP(&Content, "content", "c", "HELM UPDATE", "更新备注")
	RootCmd.PersistentFlags().StringVarP(&User, "user", "u", "Admin", "提交用户")
}
