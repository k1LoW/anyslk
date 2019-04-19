// Copyright © 2019 Ken'ichiro Oyama <k1lowxb@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/k1LoW/anyslk/logger"
	"github.com/k1LoW/anyslk/smtp_server"
	"github.com/k1LoW/anyslk/util"
	"github.com/k1LoW/anyslk/version"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	smtpPort    int
	logDir      string
	showVersion bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "anyslk",
	Short: "* -> slack message",
	Long:  `* -> slack message`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		l := logger.NewLogger(logDir)

		if showVersion {
			fmt.Printf("%s\n", version.Version)
			os.Exit(1)
		}

		_, err := util.GetEnvSlackIncommingWebhook()
		if err != nil {
			l.Fatal("error", zap.Error(err))
			os.Exit(1)
		}

		if smtpPort != 0 {
			l.Info("Starting SMTP server.")
			go smtp_server.RunServer(ctx, l, smtpPort)
		}

		signalChan := make(chan os.Signal, 1)
		signal.Ignore()
		signal.Notify(signalChan, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

		sc := <-signalChan

		switch sc {
		case syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM:
			l.Info("Shutting down servers.")
		default:
			l.Fatal("Unexpected signal.")
			os.Exit(1)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().IntVarP(&smtpPort, "smtp-port", "", 0, "SMTP server port")
	rootCmd.Flags().StringVarP(&logDir, "log-dir", "", ".", "Log directory")
	rootCmd.Flags().BoolVarP(&showVersion, "version", "v", false, "Show version")
}
