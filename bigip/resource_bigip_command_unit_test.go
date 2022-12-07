/*
Copyright 2019 F5 Networks Inc.
This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0.
If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.
*/
package bigip

import (
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/stretchr/testify/assert"
)

func TestAccBigipCommandUnitInvalid(t *testing.T) {
	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		Providers:  testProviders,
		Steps: []resource.TestStep{
			{
				Config:      testBigipCommandInvalid(),
				ExpectError: regexp.MustCompile(" Unsupported argument: An argument named \"invalidkey\" is not expected here"),
			},
		},
	})
}

func TestAccBigipCommandUnitCreate(t *testing.T) {
	setup()
	mux.HandleFunc("mgmt/shared/authn/login", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method, "Expected method 'POST', got %s", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
	})
	mux.HandleFunc("/mgmt/tm/net/self", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method, "Expected method 'GET', got %s", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		_, _ = fmt.Fprintf(w, `{}`)
	})
	mux.HandleFunc("mgmt/tm/util/bash", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method, "Expected method 'POST', got %s", r.Method)
		_, _ = fmt.Fprintf(w, `%s`, sysVersion)
	})
	//mux = http.NewServeMux()
	//mux.HandleFunc("mgmt/tm/util/bash", func(w http.ResponseWriter, r *http.Request) {
	//	_, _ = fmt.Fprintf(w, `{"kind":"tm:util:bash:runstate","command":"run","utilCmdArgs":"-c 'tmsh show sys version'"}`)
	//})
	defer teardown()
	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		Providers:  testProviders,
		Steps: []resource.TestStep{
			{
				Config: testBigipCommandCreate(server.URL),
			},
			//{
			//	Config:             testBigipCommandModify(server.URL),
			//	ExpectNonEmptyPlan: true,
			//},
		},
	})
}
func TestAccBigipCommandUnitCreateError(t *testing.T) {
	setup()
	mux.HandleFunc("mgmt/shared/authn/login", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method, "Expected method 'POST', got %s", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
	})
	mux.HandleFunc("/mgmt/tm/net/self", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method, "Expected method 'GET', got %s", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		_, _ = fmt.Fprintf(w, `{}`)
	})
	mux.HandleFunc("mgmt/tm/util/bash", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method, "Expected method 'POST', got %s", r.Method)
		http.Error(w, "The requested object name (/Common/testravi##) is invalid", http.StatusNotFound)
	})

	defer teardown()
	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		Providers:  testProviders,
		Steps: []resource.TestStep{
			{
				Config:      testBigipCommandCreate(server.URL),
				ExpectError: regexp.MustCompile("HTTP 404 :: The requested HTTP Profile \\(/Common/test-profile-http\\) was not found"),
			},
		},
	})
}

func testBigipCommandInvalid() string {
	return fmt.Sprintf(`
resource "bigip_command" "test-command" {
  commands   = ["show sys version"]
  invalidkey = "foo"
}
provider "bigip" {
  address  = "xxx.xxx.xxx.xxx"
  username = "xxx"
  password = "xxx"
}`)
}

func testBigipCommandCreate(url string) string {
	return fmt.Sprintf(`
resource "bigip_command" "test-command" {
  commands   = ["show sys version"]
}
provider "bigip" {
  address  = "%s"
  username = ""
  password = ""
  login_ref = ""
}`, url)
}

func testBigipCommandModify(url string) string {
	return fmt.Sprintf(`
resource "bigip_command" "test-command" {
  commands   = ["show sys version","show sys version"]
}
provider "bigip" {
  address  = "%s"
  username = ""
  password = ""
  login_ref = ""
}`, url)
}

var sysVersion = `{
"command": "run",
"utilCmdArgs": "-c 'tmsh show sys version'",
"commandResult": "\nSys::Version\nMain Package\n  Product     BIG-IP\n  Version     16.1.0\n  Build       0.0.19\n  Edition     Final\n  Date        Tue Jun 22 23:52:22 PDT 2021\n\n"
}`
