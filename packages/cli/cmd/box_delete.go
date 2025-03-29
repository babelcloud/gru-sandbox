package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

type BoxListResponse struct {
	Boxes []struct {
		ID string `json:"id"`
	} `json:"boxes"`
}

func NewBoxDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                "delete",
		Short:              "Delete a box by its ID",
		Long:               "Delete a box by its ID or delete all boxes",
		DisableFlagParsing: true,
		Run: func(cmd *cobra.Command, args []string) {
			var outputFormat string = "text"
			var boxID string
			var deleteAll bool
			var force bool

			// 解析参数
			for i := 0; i < len(args); i++ {
				switch args[i] {
				case "--help":
					printBoxDeleteHelp()
					return
				case "--output":
					if i+1 < len(args) {
						outputFormat = args[i+1]
						if outputFormat != "json" && outputFormat != "text" {
							fmt.Println("错误: 无效的输出格式。必须是 'json' 或 'text'")
							os.Exit(1)
						}
						i++
					} else {
						fmt.Println("错误: --output 需要参数值")
						os.Exit(1)
					}
				case "--all":
					deleteAll = true
				case "--force":
					force = true
				default:
					if !strings.HasPrefix(args[i], "-") && boxID == "" {
						boxID = args[i]
					} else if strings.HasPrefix(args[i], "-") {
						fmt.Printf("错误: 未知选项 %s\n", args[i])
						os.Exit(1)
					} else {
						fmt.Printf("错误: 意外的参数 %s\n", args[i])
						os.Exit(1)
					}
				}
			}

			// 验证参数
			if deleteAll && boxID != "" {
				fmt.Println("错误: 不能同时指定 --all 和一个盒子ID")
				os.Exit(1)
			}

			if !deleteAll && boxID == "" {
				fmt.Println("错误: 必须指定 --all 或者一个盒子ID")
				os.Exit(1)
			}

			// 处理删除所有盒子
			if deleteAll {
				// 获取所有盒子的列表
				apiURL := "http://localhost:28080/api/v1/boxes"
				if envURL := os.Getenv("API_URL"); envURL != "" {
					apiURL = envURL + "/api/v1/boxes"
				}
				resp, err := http.Get(apiURL)
				if err != nil {
					fmt.Printf("错误: 获取盒子列表失败: %v\n", err)
					os.Exit(1)
				}
				defer resp.Body.Close()

				body, err := io.ReadAll(resp.Body)
				if err != nil {
					fmt.Printf("错误: 读取响应失败: %v\n", err)
					os.Exit(1)
				}

				// 调试输出
				if os.Getenv("DEBUG") == "true" {
					fmt.Fprintf(os.Stderr, "API响应:\n")
					var prettyJSON bytes.Buffer
					if err := json.Indent(&prettyJSON, body, "", "  "); err == nil {
						fmt.Fprintln(os.Stderr, prettyJSON.String())
					} else {
						fmt.Fprintln(os.Stderr, string(body))
					}
				}

				var response BoxListResponse
				if err := json.Unmarshal(body, &response); err != nil {
					fmt.Printf("错误: 解析JSON响应失败: %v\n", err)
					os.Exit(1)
				}

				if len(response.Boxes) == 0 {
					if outputFormat == "json" {
						fmt.Println(`{"status":"success","message":"没有盒子需要删除"}`)
					} else {
						fmt.Println("没有盒子需要删除")
					}
					return
				}

				// 显示将要删除的盒子
				fmt.Println("以下盒子将被删除:")
				for _, box := range response.Boxes {
					fmt.Printf("  - %s\n", box.ID)
				}
				fmt.Println()

				// 如果非强制，确认删除
				if !force {
					fmt.Print("您确定要删除所有盒子吗? [y/N] ")
					reader := bufio.NewReader(os.Stdin)
					reply, err := reader.ReadString('\n')
					if err != nil {
						fmt.Printf("错误: 读取输入失败: %v\n", err)
						os.Exit(1)
					}

					reply = strings.TrimSpace(strings.ToLower(reply))
					if reply != "y" && reply != "yes" {
						if outputFormat == "json" {
							fmt.Println(`{"status":"cancelled","message":"操作被用户取消"}`)
						} else {
							fmt.Println("操作已取消")
						}
						return
					}
				}

				// 删除所有盒子
				success := true
				for _, box := range response.Boxes {
					apiURL := fmt.Sprintf("http://localhost:28080/api/v1/boxes/%s", box.ID)
					if envURL := os.Getenv("API_URL"); envURL != "" {
						apiURL = fmt.Sprintf("%s/api/v1/boxes/%s", envURL, box.ID)
					}
					req, err := http.NewRequest("DELETE", apiURL, strings.NewReader(`{"force":true}`))
					if err != nil {
						fmt.Printf("错误: 创建请求失败: %v\n", err)
						success = false
						continue
					}
					req.Header.Set("Content-Type", "application/json")

					client := &http.Client{}
					resp, err := client.Do(req)
					if err != nil {
						fmt.Printf("错误: 删除盒子 %s 失败: %v\n", box.ID, err)
						success = false
						continue
					}
					resp.Body.Close()

					if resp.StatusCode != 200 && resp.StatusCode != 204 {
						fmt.Printf("错误: 删除盒子 %s 失败，HTTP状态码: %d\n", box.ID, resp.StatusCode)
						success = false
					}
				}

				if success {
					if outputFormat == "json" {
						fmt.Println(`{"status":"success","message":"所有盒子删除成功"}`)
					} else {
						fmt.Println("所有盒子删除成功")
					}
				} else {
					if outputFormat == "json" {
						fmt.Println(`{"status":"error","message":"一些盒子删除失败"}`)
					} else {
						fmt.Println("一些盒子删除失败")
					}
					os.Exit(1)
				}
				return
			}

			// 删除单个盒子
			apiURL := fmt.Sprintf("http://localhost:28080/api/v1/boxes/%s", boxID)
			if envURL := os.Getenv("API_URL"); envURL != "" {
				apiURL = fmt.Sprintf("%s/api/v1/boxes/%s", envURL, boxID)
			}
			req, err := http.NewRequest("DELETE", apiURL, strings.NewReader(`{"force":true}`))
			if err != nil {
				fmt.Printf("错误: 创建请求失败: %v\n", err)
				os.Exit(1)
			}
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				fmt.Printf("错误: 删除盒子失败。确保API服务器正在运行且ID '%s' 正确\n", boxID)
				if os.Getenv("DEBUG") == "true" {
					fmt.Fprintf(os.Stderr, "错误: %v\n", err)
				}
				os.Exit(1)
			}
			defer resp.Body.Close()

			if resp.StatusCode != 200 && resp.StatusCode != 204 {
				fmt.Printf("错误: 删除盒子失败，HTTP状态码: %d\n", resp.StatusCode)
				body, _ := io.ReadAll(resp.Body)
				if os.Getenv("DEBUG") == "true" && len(body) > 0 {
					fmt.Fprintf(os.Stderr, "响应: %s\n", string(body))
				}
				os.Exit(1)
			}

			if outputFormat == "json" {
				fmt.Println(`{"status":"success","message":"盒子删除成功"}`)
			} else {
				fmt.Println("盒子删除成功")
			}
		},
	}

	return cmd
}

func printBoxDeleteHelp() {
	fmt.Println("用法: gbox box delete [选项] <id>")
	fmt.Println()
	fmt.Println("选项:")
	fmt.Println("    --output          输出格式 (json或text, 默认: text)")
	fmt.Println("    --all             删除所有盒子")
	fmt.Println("    --force           强制删除，无需确认")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("    gbox box delete 550e8400-e29b-41d4-a716-446655440000              # 删除一个盒子")
	fmt.Println("    gbox box delete --all --force                                     # 无需确认删除所有盒子")
	fmt.Println("    gbox box delete --all                                             # 删除所有盒子(需确认)")
	fmt.Println("    gbox box delete 550e8400-e29b-41d4-a716-446655440000 --output json  # 删除盒子并输出JSON")
	fmt.Println()
}
