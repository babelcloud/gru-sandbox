package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

// BoxPath 表示盒子路径的结构
type BoxPath struct {
	BoxID string
	Path  string
}

// 解析盒子路径（格式为 BOX_ID:PATH）
func parseBoxPath(path string) (*BoxPath, error) {
	re := regexp.MustCompile(`^([^:]+):(.+)$`)
	matches := re.FindStringSubmatch(path)
	if len(matches) != 3 {
		return nil, fmt.Errorf("无效的盒子路径格式，应为 BOX_ID:PATH")
	}
	return &BoxPath{
		BoxID: matches[1],
		Path:  matches[2],
	}, nil
}

// 检查路径是否为盒子路径
func isBoxPath(path string) bool {
	return strings.Contains(path, ":")
}

// 将相对路径转换为绝对路径
func getAbsolutePath(path string) string {
	if _, err := os.Stat(path); err == nil {
		absPath, err := filepath.Abs(path)
		if err == nil {
			return absPath
		}
	}

	dir := filepath.Dir(path)
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return path
	}

	return filepath.Join(absDir, filepath.Base(path))
}

func NewBoxCpCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                "cp",
		Short:              "Copy files/folders between a box and the local filesystem",
		Long:               "Copy files/folders between a box and the local filesystem",
		DisableFlagParsing: true,
		Run: func(cmd *cobra.Command, args []string) {
			// 帮助信息
			if len(args) == 1 && (args[0] == "--help" || args[0] == "help") {
				printBoxCpHelp()
				return
			}

			// 参数验证
			if len(args) != 2 {
				printBoxCpHelp()
				os.Exit(1)
			}

			src := args[0]
			dst := args[1]
			debugEnabled := os.Getenv("DEBUG") == "true"
			apiURL := "http://localhost:28080/api/v1"
			if envURL := os.Getenv("API_URL"); envURL != "" {
				apiURL = envURL + "/api/v1"
			}

			// 调试日志
			debug := func(msg string) {
				if debugEnabled {
					fmt.Fprintf(os.Stderr, "[DEBUG] %s\n", msg)
				}
			}

			// 判断复制方向并处理
			if isBoxPath(src) && !isBoxPath(dst) {
				// 从盒子复制到本地
				boxPath, err := parseBoxPath(src)
				if err != nil {
					fmt.Println("错误: ", err)
					os.Exit(1)
				}

				debug(fmt.Sprintf("盒子ID: %s", boxPath.BoxID))
				debug(fmt.Sprintf("源路径: %s", boxPath.Path))
				debug(fmt.Sprintf("目标路径: %s", dst))

				if dst == "-" {
					// 从盒子复制到标准输出作为tar流
					requestURL := fmt.Sprintf("%s/boxes/%s/archive?path=%s", apiURL, boxPath.BoxID, boxPath.Path)
					debug(fmt.Sprintf("发送GET请求到: %s", requestURL))

					resp, err := http.Get(requestURL)
					if err != nil {
						fmt.Println("错误: 从盒子下载失败")
						os.Exit(1)
					}
					defer resp.Body.Close()

					debug(fmt.Sprintf("HTTP响应状态码: %d", resp.StatusCode))

					if resp.StatusCode != http.StatusOK {
						fmt.Println("错误: 从盒子下载失败，HTTP状态码:", resp.StatusCode)
						os.Exit(1)
					}

					_, err = io.Copy(os.Stdout, resp.Body)
					if err != nil {
						fmt.Println("错误: 写入标准输出失败")
						os.Exit(1)
					}
				} else {
					// 将本地路径转换为绝对路径
					dst = getAbsolutePath(dst)
					debug(fmt.Sprintf("绝对目标路径: %s", dst))

					// 复制从盒子到本地文件
					err := os.MkdirAll(filepath.Dir(dst), 0755)
					if err != nil {
						fmt.Printf("错误: 创建目标目录失败: %v\n", err)
						os.Exit(1)
					}

					// 下载到临时文件
					tempFile, err := os.CreateTemp("", "gbox-cp-")
					if err != nil {
						fmt.Printf("错误: 创建临时文件失败: %v\n", err)
						os.Exit(1)
					}
					tempFilePath := tempFile.Name()
					defer os.Remove(tempFilePath)

					requestURL := fmt.Sprintf("%s/boxes/%s/archive?path=%s", apiURL, boxPath.BoxID, boxPath.Path)
					debug(fmt.Sprintf("发送GET请求到: %s", requestURL))

					resp, err := http.Get(requestURL)
					if err != nil {
						fmt.Println("错误: 从盒子下载失败")
						os.Exit(1)
					}
					defer resp.Body.Close()

					debug(fmt.Sprintf("HTTP响应状态码: %d", resp.StatusCode))

					if resp.StatusCode != http.StatusOK {
						fmt.Println("错误: 从盒子下载失败，HTTP状态码:", resp.StatusCode)
						os.Exit(1)
					}

					_, err = io.Copy(tempFile, resp.Body)
					tempFile.Close()
					if err != nil {
						fmt.Printf("错误: 写入临时文件失败: %v\n", err)
						os.Exit(1)
					}

					// 检查格式并解压
					debug(fmt.Sprintf("解压归档文件到: %s", filepath.Dir(dst)))
					dstDir := filepath.Dir(dst)
					srcBaseName := filepath.Base(boxPath.Path)

					// 尝试作为gzip tar解压
					cmd := exec.Command("tar", "-xzf", tempFilePath, "-C", dstDir, srcBaseName)
					err = cmd.Run()
					if err != nil {
						// 尝试作为普通tar解压
						cmd = exec.Command("tar", "-xf", tempFilePath, "-C", dstDir, srcBaseName)
						err = cmd.Run()
						if err != nil {
							fmt.Printf("错误: 解压归档文件失败: %v\n", err)
							os.Exit(1)
						}
					}

					fmt.Fprintf(os.Stderr, "已复制从盒子 %s:%s 到 %s\n", boxPath.BoxID, boxPath.Path, dst)
				}
			} else if !isBoxPath(src) && isBoxPath(dst) {
				// 从本地复制到盒子
				boxPath, err := parseBoxPath(dst)
				if err != nil {
					fmt.Println("错误: ", err)
					os.Exit(1)
				}

				debug(fmt.Sprintf("盒子ID: %s", boxPath.BoxID))
				debug(fmt.Sprintf("目标路径: %s", boxPath.Path))
				debug(fmt.Sprintf("源路径: %s", src))

				if src == "-" {
					// 从标准输入复制tar流到盒子
					requestURL := fmt.Sprintf("%s/boxes/%s/archive?path=%s", apiURL, boxPath.BoxID, boxPath.Path)
					debug(fmt.Sprintf("发送PUT请求到: %s", requestURL))

					req, err := http.NewRequest("PUT", requestURL, os.Stdin)
					if err != nil {
						fmt.Printf("错误: 创建请求失败: %v\n", err)
						os.Exit(1)
					}

					req.Header.Set("Content-Type", "application/x-tar")

					client := &http.Client{}
					resp, err := client.Do(req)
					if err != nil {
						fmt.Println("错误: 上传到盒子失败")
						os.Exit(1)
					}
					defer resp.Body.Close()

					if resp.StatusCode != http.StatusOK {
						fmt.Println("错误: 上传到盒子失败，HTTP状态码:", resp.StatusCode)
						os.Exit(1)
					}

					fmt.Fprintf(os.Stderr, "已复制从标准输入到盒子 %s:%s\n", boxPath.BoxID, boxPath.Path)
				} else {
					// 将本地路径转换为绝对路径
					src = getAbsolutePath(src)
					debug(fmt.Sprintf("绝对源路径: %s", src))

					// 检查源文件是否存在
					if _, err := os.Stat(src); os.IsNotExist(err) {
						fmt.Printf("错误: 源文件或目录不存在: %s\n", src)
						os.Exit(1)
					}

					// 复制从本地文件到盒子
					tempFile, err := os.CreateTemp("", "gbox-cp-")
					if err != nil {
						fmt.Printf("错误: 创建临时文件失败: %v\n", err)
						os.Exit(1)
					}
					tempFilePath := tempFile.Name()
					tempFile.Close()
					defer os.Remove(tempFilePath)

					// 创建tar归档
					cmd := exec.Command("tar", "--no-xattrs", "-czf", tempFilePath, "-C", filepath.Dir(src), filepath.Base(src))
					err = cmd.Run()
					if err != nil {
						fmt.Printf("错误: 创建tar归档失败: %v\n", err)
						os.Exit(1)
					}
					debug(fmt.Sprintf("已创建tar归档: %s", src))

					// 获取文件大小
					fileInfo, err := os.Stat(tempFilePath)
					if err != nil {
						fmt.Printf("错误: 获取临时文件信息失败: %v\n", err)
						os.Exit(1)
					}
					fileSize := fileInfo.Size()

					// 上传归档到盒子
					file, err := os.Open(tempFilePath)
					if err != nil {
						fmt.Printf("错误: 打开临时文件失败: %v\n", err)
						os.Exit(1)
					}
					defer file.Close()

					requestURL := fmt.Sprintf("%s/boxes/%s/archive?path=%s", apiURL, boxPath.BoxID, boxPath.Path)
					debug(fmt.Sprintf("发送PUT请求到: %s", requestURL))

					req, err := http.NewRequest("PUT", requestURL, file)
					if err != nil {
						fmt.Printf("错误: 创建请求失败: %v\n", err)
						os.Exit(1)
					}

					req.Header.Set("Content-Type", "application/x-tar")
					req.Header.Set("Content-Length", fmt.Sprintf("%d", fileSize))

					client := &http.Client{}
					resp, err := client.Do(req)
					if err != nil {
						fmt.Println("错误: 上传到盒子失败")
						os.Exit(1)
					}
					defer resp.Body.Close()

					if resp.StatusCode != http.StatusOK {
						fmt.Println("错误: 上传到盒子失败，HTTP状态码:", resp.StatusCode)
						os.Exit(1)
					}

					fmt.Fprintf(os.Stderr, "已复制从 %s 到盒子 %s:%s\n", src, boxPath.BoxID, boxPath.Path)
				}
			} else {
				fmt.Println("错误: 无效的路径格式。一个路径必须是盒子路径(BOX_ID:PATH)，另一个必须是本地路径")
				os.Exit(1)
			}
		},
	}

	return cmd
}

func printBoxCpHelp() {
	fmt.Println("用法: gbox box cp <src> <dst>")
	fmt.Println()
	fmt.Println("参数:")
	fmt.Println("    <src>  源路径。可以是:")
	fmt.Println("           - 本地文件/目录路径 (例如，./local_file, /tmp/data)")
	fmt.Println("           - 盒子路径，格式为 BOX_ID:SRC_PATH (例如，550e8400-e29b-41d4-a716-446655440000:/work)")
	fmt.Println("           - \"-\" 表示从标准输入读取 (必须是tar流)")
	fmt.Println()
	fmt.Println("    <dst>  目标路径。可以是:")
	fmt.Println("           - 本地文件/目录路径 (例如，/tmp/app_logs)")
	fmt.Println("           - 盒子路径，格式为 BOX_ID:DST_PATH (例如，550e8400-e29b-41d4-a716-446655440000:/work)")
	fmt.Println("           - \"-\" 表示写入标准输出 (作为tar流)")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("    # 复制本地文件到盒子")
	fmt.Println("    gbox box cp ./local_file 550e8400-e29b-41d4-a716-446655440000:/work")
	fmt.Println()
	fmt.Println("    # 从盒子复制到本地")
	fmt.Println("    gbox box cp 550e8400-e29b-41d4-a716-446655440000:/var/logs/ /tmp/app_logs")
	fmt.Println()
	fmt.Println("    # 从标准输入复制tar流到盒子")
	fmt.Println("    tar czf - ./local_dir | gbox box cp - 550e8400-e29b-41d4-a716-446655440000:/work")
	fmt.Println()
	fmt.Println("    # 从盒子复制到标准输出作为tar流")
	fmt.Println("    gbox box cp 550e8400-e29b-41d4-a716-446655440000:/etc/hosts - | tar xzf -")
	fmt.Println()
	fmt.Println("    # 复制目录从本地到盒子")
	fmt.Println("    gbox box cp ./app_data 550e8400-e29b-41d4-a716-446655440000:/data/")
	fmt.Println()
	fmt.Println("    # 复制目录从盒子到本地")
	fmt.Println("    gbox box cp 550e8400-e29b-41d4-a716-446655440000:/var/logs/ /tmp/app_logs/")
}

// This is a placeholder file
// TODO: Implement box_cp functionality
