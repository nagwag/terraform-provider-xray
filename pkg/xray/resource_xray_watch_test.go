package xray

import (
	"fmt"
	"github.com/jfrog/terraform-provider-shared/test"
	"github.com/jfrog/terraform-provider-shared/util"
	"regexp"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/jfrog/terraform-provider-shared/client"
)

var testDataWatch = map[string]string{
	"resource_name":     "",
	"watch_name":        "xray-watch",
	"description":       "This is a new watch created by TF Provider",
	"active":            "true",
	"watch_type":        "all-repos",
	"filter_type_0":     "regex",
	"filter_value_0":    ".*",
	"filter_type_1":     "package-type",
	"filter_value_1":    "Docker",
	"policy_name_0":     "xray-policy-0",
	"policy_name_1":     "xray-policy-1",
	"watch_recipient_0": "test@email.com",
	"watch_recipient_1": "test@email.com",
}

func TestAccWatch_allReposSinglePolicy(t *testing.T) {
	_, fqrn, resourceName := test.MkNames("watch-", "xray_watch")
	testData := util.MergeMaps(testDataWatch)

	testData["resource_name"] = resourceName
	testData["watch_name"] = fmt.Sprintf("xray-watch-%d", test.RandomInt())
	testData["policy_name_0"] = fmt.Sprintf("xray-policy-%d", test.RandomInt())

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		CheckDestroy: verifyDeleted(fqrn, func(id string, request *resty.Request) (*resty.Response, error) {
			testCheckPolicyDeleted("xray_security_policy.security", t, request)
			resp, err := testCheckWatch(id, request)
			return resp, err
		}),
		ProviderFactories: testAccProviders(),
		Steps: []resource.TestStep{
			{
				Config: util.ExecuteTemplate(fqrn, allReposSinglePolicyWatchTemplate, testData),
				Check:  verifyXrayWatch(fqrn, testData),
			},
		},
	})
}

func TestAccWatch_allReposWithProjectKey(t *testing.T) {
	_, fqrn, resourceName := test.MkNames("watch-", "xray_watch")
	projectKey := fmt.Sprintf("testproj%d", test.RandSelect(1, 2, 3, 4, 5))

	testData := util.MergeMaps(testDataWatch)

	testData["resource_name"] = resourceName
	testData["project_key"] = projectKey
	testData["watch_name"] = fmt.Sprintf("xray-watch-%d", test.RandomInt())
	testData["policy_name_0"] = fmt.Sprintf("xray-policy-%d", test.RandomInt())

	template := `resource "xray_security_policy" "security" {
	  name        = "{{ .policy_name_0 }}"
	  description = "Security policy description"
	  type        = "security"
	  rule {
		name     = "rule-name-severity"
		priority = 1
		criteria {
		  min_severity = "High"
		}
		actions {
		  webhooks = []
		  mails    = ["test@email.com"]
		  block_download {
			unscanned = true
			active    = true
		  }
		  block_release_bundle_distribution  = true
		  fail_build                         = true
		  notify_watch_recipients            = true
		  notify_deployer                    = true
		  create_ticket_enabled              = false
		  build_failure_grace_period_in_days = 5
		}
	  }
	}

	resource "xray_watch" "{{ .resource_name }}" {
	  name        	= "{{ .watch_name }}"
	  description 	= "{{ .description }}"
	  active 		= {{ .active }}
	  project_key   = "{{ .project_key }}"

	  watch_resource {
		type       	= "{{ .watch_type }}"
		filter {
			type  	= "{{ .filter_type_0 }}"
			value	= "{{ .filter_value_0 }}"
		}
	  }
	  assigned_policy {
		name 	= xray_security_policy.security.name
		type 	= "security"
	  }

	  watch_recipients = ["{{ .watch_recipient_0 }}", "{{ .watch_recipient_1 }}"]
	}`
	config := util.ExecuteTemplate(fqrn, template, testData)

	updatedTestData := util.MergeMaps(testData)
	updatedTestData["description"] = "New description"
	updatedConfig := util.ExecuteTemplate(fqrn, template, updatedTestData)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			CreateProject(t, projectKey)
		},
		CheckDestroy: verifyDeleted(fqrn, func(id string, request *resty.Request) (*resty.Response, error) {
			DeleteProject(t, projectKey)
			resp, err := testCheckWatch(id, request)
			return resp, err
		}),
		ProviderFactories: testAccProviders(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					verifyXrayWatch(fqrn, testData),
					resource.TestCheckResourceAttr(fqrn, "project_key", projectKey),
				),
			},
			{
				Config: updatedConfig,
				Check:  verifyXrayWatch(fqrn, updatedTestData),
			},
		},
	})
}

func TestAccWatch_allReposMultiplePolicies(t *testing.T) {
	_, fqrn, resourceName := test.MkNames("watch-", "xray_watch")
	testData := util.MergeMaps(testDataWatch)

	testData["resource_name"] = resourceName
	testData["watch_name"] = fmt.Sprintf("xray-watch-%d", test.RandomInt())
	testData["policy_name_0"] = fmt.Sprintf("xray-policy-1%d", test.RandomInt())
	testData["policy_name_1"] = fmt.Sprintf("xray-policy-2%d", test.RandomInt())

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		CheckDestroy: verifyDeleted(fqrn, func(id string, request *resty.Request) (*resty.Response, error) {
			testCheckPolicyDeleted("xray_security_policy.security", t, request)
			testCheckPolicyDeleted("xray_license_policy.license", t, request)
			resp, err := testCheckWatch(id, request)
			return resp, err
		}),

		ProviderFactories: testAccProviders(),
		Steps: []resource.TestStep{
			{
				Config: util.ExecuteTemplate(fqrn, allReposMultiplePoliciesWatchTemplate, testData),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", testData["watch_name"]),
					resource.TestCheckResourceAttr(fqrn, "description", testData["description"]),
					resource.TestCheckResourceAttr(fqrn, "watch_resource.0.type", testData["watch_type"]),
					resource.TestCheckTypeSetElemNestedAttrs(fqrn, "watch_resource.0.filter.*", map[string]string{
						"type":  "regex",
						"value": ".*",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(fqrn, "watch_resource.0.filter.*", map[string]string{
						"type":  "package-type",
						"value": "Docker",
					}),
					resource.TestCheckResourceAttr(fqrn, "assigned_policy.0.name", testData["policy_name_0"]),
					resource.TestCheckResourceAttr(fqrn, "assigned_policy.0.type", "security"),
					resource.TestCheckResourceAttr(fqrn, "assigned_policy.1.name", testData["policy_name_1"]),
					resource.TestCheckResourceAttr(fqrn, "assigned_policy.1.type", "license"),
				),
			},
		},
	})
}

func makeSingleRepositoryTestCase(repoType string, t *testing.T) (*testing.T, resource.TestCase) {
	_, fqrn, resourceName := test.MkNames("watch-", "xray_watch")
	testData := util.MergeMaps(testDataWatch)

	testData["resource_name"] = resourceName
	testData["watch_name"] = fmt.Sprintf("xray-watch-%d", test.RandomInt())
	testData["policy_name_0"] = fmt.Sprintf("xray-policy-%d", test.RandomInt())
	testData["watch_type"] = "repository"
	testData["repo_type"] = repoType
	testData["repo0"] = fmt.Sprintf("libs-release-%s-0-%d", repoType, test.RandomInt())

	return t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCreateRepos(t, testData["repo0"], repoType, "")
		},
		CheckDestroy: verifyDeleted(fqrn, func(id string, request *resty.Request) (*resty.Response, error) {
			testAccDeleteRepo(t, testData["repo0"])
			testCheckPolicyDeleted("xray_security_policy.security", t, request)
			resp, err := testCheckWatch(id, request)
			return resp, err
		}),

		ProviderFactories: testAccProviders(),
		Steps: []resource.TestStep{
			{
				Config: util.ExecuteTemplate(fqrn, singleRepositoryWatchTemplate, testData),
				Check: resource.ComposeTestCheckFunc(
					verifyXrayWatch(fqrn, testData),
					resource.TestCheckTypeSetElemNestedAttrs(fqrn, "watch_resource.*.filter.*", map[string]string{
						"type":  "regex",
						"value": ".*",
					}),
				),
			},
		},
	}
}

// To verify the watch for a single repo we need to create a new repository with Xray indexing enabled
// testAccCreateRepos() is creating a repos with Xray indexing enabled using the API call
// We need to figure out how to use external providers (like Artifactory) in the tests. Documented approach didn't work
func TestAccWatch_singleRepository(t *testing.T) {
	for _, repoType := range []string{"local", "remote"} {
		t.Run(repoType, func(t *testing.T) {
			resource.Test(makeSingleRepositoryTestCase(repoType, t))
		})
	}
}

func TestAccWatch_singleRepositoryWithProjectKey(t *testing.T) {
	_, fqrn, resourceName := test.MkNames("watch-", "xray_watch")
	repoKey := fmt.Sprintf("local-%d", test.RandomInt())
	projectKey := fmt.Sprintf("testproj%d", test.RandSelect(1, 2, 3, 4, 5))

	testData := util.MergeMaps(testDataWatch)
	testData["resource_name"] = resourceName
	testData["project_key"] = projectKey
	testData["watch_name"] = fmt.Sprintf("xray-watch-%d", test.RandomInt())
	testData["policy_name_0"] = fmt.Sprintf("xray-policy-%d", test.RandomInt())
	testData["watch_type"] = "repository"
	testData["repo_type"] = "local"
	testData["repo0"] = repoKey

	template := `resource "xray_security_policy" "security" {
	  name        = "{{ .policy_name_0 }}"
	  description = "Security policy description"
	  type        = "security"
	  rule {
	    name     = "rule-name-severity"
	    priority = 1
	    criteria {
	      min_severity = "High"
	    }
	    actions {
	      webhooks = []
	      mails    = ["test@email.com"]
	      block_download {
	        unscanned = true
	        active    = true
	      }
	      block_release_bundle_distribution  = true
	      fail_build                         = true
	      notify_watch_recipients            = true
	      notify_deployer                    = true
	      create_ticket_enabled              = false
	      build_failure_grace_period_in_days = 5
	    }
	  }
	}

	resource "xray_watch" "{{ .resource_name }}" {
	  name        	= "{{ .watch_name }}"
	  description 	= "{{ .description }}"
	  active 		= {{ .active }}
	  project_key   = "{{ .project_key }}"

	  watch_resource {
		type       	= "{{ .watch_type }}"
		bin_mgr_id  = "default"
		name		= "{{ .repo0 }}"
		repo_type   = "{{ .repo_type }}"
	  }
	  assigned_policy {
	  	name 	= xray_security_policy.security.name
	  	type 	= "security"
	  }
	  watch_recipients = ["{{ .watch_recipient_0 }}", "{{ .watch_recipient_1 }}"]
	}`

	config := util.ExecuteTemplate(fqrn, template, testData)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			CreateProject(t, projectKey)
			testAccCreateRepos(t, repoKey, "local", projectKey)
		},
		CheckDestroy: verifyDeleted(fqrn, func(id string, request *resty.Request) (*resty.Response, error) {
			testAccDeleteRepo(t, repoKey)
			DeleteProject(t, projectKey)
			resp, err := testCheckWatch(id, request)
			return resp, err
		}),

		ProviderFactories: testAccProviders(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					verifyXrayWatch(fqrn, testData),
					resource.TestCheckResourceAttr(fqrn, "project_key", projectKey),
				),
			},
		},
	})
}

func TestAccWatch_repositoryMissingRepoType(t *testing.T) {
	_, fqrn, resourceName := test.MkNames("watch-", "xray_watch")
	testData := util.MergeMaps(testDataWatch)

	testData["resource_name"] = resourceName
	testData["watch_name"] = fmt.Sprintf("xray-watch-%d", test.RandomInt())
	testData["policy_name_0"] = fmt.Sprintf("xray-policy-%d", test.RandomInt())
	testData["watch_type"] = "repository"
	testData["repo0"] = fmt.Sprintf("libs-release-local-0-%d", test.RandomInt())

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCreateRepos(t, testData["repo0"], "local", "")
		},
		CheckDestroy: verifyDeleted(fqrn, func(id string, request *resty.Request) (*resty.Response, error) {
			testAccDeleteRepo(t, testData["repo0"])
			testCheckPolicyDeleted("xray_security_policy.security", t, request)
			resp, err := testCheckWatch(id, request)
			return resp, err
		}),

		ProviderFactories: testAccProviders(),
		Steps: []resource.TestStep{
			{
				Config:      util.ExecuteTemplate(fqrn, singleRepositoryInvalidWatchTemplate, testData),
				ExpectError: regexp.MustCompile(`attribute 'repo_type' not set when 'watch_resource\.type' is set to 'repository'`),
			},
		},
	})
}

func TestAccWatch_multipleRepositories(t *testing.T) {
	_, fqrn, resourceName := test.MkNames("watch-", "xray_watch")
	testData := util.MergeMaps(testDataWatch)

	testData["resource_name"] = resourceName
	testData["watch_name"] = fmt.Sprintf("xray-watch-%d", test.RandomInt())
	testData["policy_name_0"] = fmt.Sprintf("xray-policy-%d", test.RandomInt())
	testData["watch_type"] = "repository"
	testData["repo_type"] = "local"
	testData["repo0"] = fmt.Sprintf("libs-release-local-0-%d", test.RandomInt())
	testData["repo1"] = fmt.Sprintf("libs-release-local-1-%d", test.RandomInt())

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCreateRepos(t, testData["repo0"], "local", "")
			testAccCreateRepos(t, testData["repo1"], "local", "")
		},
		CheckDestroy: verifyDeleted(fqrn, func(id string, request *resty.Request) (*resty.Response, error) {
			testAccDeleteRepo(t, testData["repo0"])
			testAccDeleteRepo(t, testData["repo1"])
			testCheckPolicyDeleted("xray_security_policy.security", t, request)
			resp, err := testCheckWatch(id, request)
			return resp, err
		}),
		ProviderFactories: testAccProviders(),
		Steps: []resource.TestStep{
			{
				Config: util.ExecuteTemplate(fqrn, multipleRepositoriesWatchTemplate, testData),
				Check:  verifyXrayWatch(fqrn, testData),
			},
		},
	})
}

func TestAccWatch_build(t *testing.T) {
	_, fqrn, resourceName := test.MkNames("watch-", "xray_watch")
	testData := util.MergeMaps(testDataWatch)

	testData["resource_name"] = resourceName
	testData["watch_name"] = fmt.Sprintf("xray-watch-%d", test.RandomInt())
	testData["policy_name_0"] = fmt.Sprintf("xray-policy-%d", test.RandomInt())
	testData["watch_type"] = "build"
	testData["build_name0"] = fmt.Sprintf("release-pipeline-%d", test.RandomInt())
	testData["build_name1"] = fmt.Sprintf("release-pipeline1-%d", test.RandomInt())
	builds := []string{testData["build_name0"]}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCreateBuilds(t, builds, "")
		},
		CheckDestroy:      verifyDeleted(fqrn, testCheckWatch),
		ProviderFactories: testAccProviders(),
		Steps: []resource.TestStep{
			{
				Config: util.ExecuteTemplate(fqrn, buildWatchTemplate, testData),
				Check:  verifyXrayWatch(fqrn, testData),
			},
		},
	})
}

func TestAccWatch_buildWithProjectKey(t *testing.T) {
	_, fqrn, resourceName := test.MkNames("watch-", "xray_watch")
	projectKey := fmt.Sprintf("testproj%d", test.RandSelect(1, 2, 3, 4, 5))

	testData := util.MergeMaps(testDataWatch)
	testData["resource_name"] = resourceName
	testData["project_key"] = projectKey
	testData["watch_name"] = fmt.Sprintf("xray-watch-%d", test.RandomInt())
	testData["policy_name_0"] = fmt.Sprintf("xray-policy-%d", test.RandomInt())
	testData["watch_type"] = "build"
	testData["build_name0"] = fmt.Sprintf("release-pipeline-%d", test.RandomInt())

	template := `resource "xray_security_policy" "security" {
	  name        = "{{ .policy_name_0 }}"
	  description = "Security policy description"
	  type        = "security"
	  rule {
	    name     = "rule-name-severity"
	    priority = 1
	    criteria {
	      min_severity = "High"
	    }
	    actions {
	      webhooks = []
	      mails    = ["test@email.com"]
	      block_download {
	        unscanned = true
	        active    = true
	      }
	      block_release_bundle_distribution  = true
	      fail_build                         = true
	      notify_watch_recipients            = true
	      notify_deployer                    = true
	      create_ticket_enabled              = false
	      build_failure_grace_period_in_days = 5
	    }
	  }
	}

	resource "xray_watch" "{{ .resource_name }}" {
	  name        	= "{{ .watch_name }}"
	  description 	= "{{ .description }}"
	  active 		= {{ .active }}
	  project_key   = "{{ .project_key }}"

	  watch_resource {
		type       	= "{{ .watch_type }}"
		bin_mgr_id  = "default"
		name		= "{{ .build_name0 }}"
	  }
	  assigned_policy {
	  	name 	= xray_security_policy.security.name
	  	type 	= "security"
	  }
	  watch_recipients = ["{{ .watch_recipient_0 }}", "{{ .watch_recipient_1 }}"]
	}`
	config := util.ExecuteTemplate(fqrn, template, testData)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			CreateProject(t, projectKey)
			testAccCreateBuilds(t, []string{testData["build_name0"]}, projectKey)
		},
		CheckDestroy: verifyDeleted(fqrn, func(id string, request *resty.Request) (*resty.Response, error) {
			DeleteProject(t, projectKey)
			resp, err := testCheckWatch(id, request)
			return resp, err
		}),
		ProviderFactories: testAccProviders(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					verifyXrayWatch(fqrn, testData),
					resource.TestCheckResourceAttr(fqrn, "project_key", projectKey),
				),
			},
		},
	})
}

func TestAccWatch_allBuildsWithProjectKey(t *testing.T) {
	_, fqrn, resourceName := test.MkNames("watch-", "xray_watch")
	projectKey := fmt.Sprintf("testproj%d", test.RandSelect(1, 2, 3, 4, 5))

	testData := util.MergeMaps(testDataWatch)
	testData["resource_name"] = resourceName
	testData["project_key"] = projectKey
	testData["watch_name"] = fmt.Sprintf("xray-watch-%d", test.RandomInt())
	testData["policy_name_0"] = fmt.Sprintf("xray-policy-%d", test.RandomInt())
	testData["watch_type"] = "all-builds"

	template := `resource "xray_security_policy" "security" {
	  name        = "{{ .policy_name_0 }}"
	  description = "Security policy description"
	  type        = "security"
	  rule {
	    name     = "rule-name-severity"
	    priority = 1
	    criteria {
	      min_severity = "High"
	    }
	    actions {
	      webhooks = []
	      mails    = ["test@email.com"]
	      block_download {
	        unscanned = true
	        active    = true
	      }
	      block_release_bundle_distribution  = true
	      fail_build                         = true
	      notify_watch_recipients            = true
	      notify_deployer                    = true
	      create_ticket_enabled              = false
	      build_failure_grace_period_in_days = 5
	    }
	  }
	}

	resource "xray_watch" "{{ .resource_name }}" {
	  name        	= "{{ .watch_name }}"
	  description 	= "{{ .description }}"
	  active 		= {{ .active }}
	  project_key   = "{{ .project_key }}"

	  watch_resource {
		type       	= "{{ .watch_type }}"
		bin_mgr_id  = "default"
		ant_filter {
			exclude_patterns = ["a*", "b*"]
			include_patterns = ["ab*"]
		}
		ant_filter {
			exclude_patterns = ["c*", "d*"]
			include_patterns = ["cd*"]
		}
	  }

	  assigned_policy {
	  	name 	= xray_security_policy.security.name
	  	type 	= "security"
	  }
	  watch_recipients = ["{{ .watch_recipient_0 }}", "{{ .watch_recipient_1 }}"]
	}`
	config := util.ExecuteTemplate(fqrn, template, testData)

	updatedTestData := util.MergeMaps(testData)
	updatedTestData["description"] = "New description"
	updatedConfig := util.ExecuteTemplate(fqrn, template, updatedTestData)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			CreateProject(t, projectKey)
		},
		CheckDestroy: verifyDeleted(fqrn, func(id string, request *resty.Request) (*resty.Response, error) {
			DeleteProject(t, projectKey)
			resp, err := testCheckWatch(id, request)
			return resp, err
		}),
		ProviderFactories: testAccProviders(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					verifyXrayWatch(fqrn, testData),
					resource.TestCheckResourceAttr(fqrn, "project_key", projectKey),
				),
			},
			{
				Config: updatedConfig,
				Check:  verifyXrayWatch(fqrn, updatedTestData),
			},
		},
	})
}

func TestAccWatch_multipleBuilds(t *testing.T) {
	_, fqrn, resourceName := test.MkNames("watch-", "xray_watch")
	testData := util.MergeMaps(testDataWatch)

	testData["resource_name"] = resourceName
	testData["watch_name"] = fmt.Sprintf("xray-watch-%d", test.RandomInt())
	testData["policy_name_0"] = fmt.Sprintf("xray-policy-%d", test.RandomInt())
	testData["watch_type"] = "build"
	testData["build_name0"] = fmt.Sprintf("release-pipeline-%d", test.RandomInt())
	testData["build_name1"] = fmt.Sprintf("release-pipeline1-%d", test.RandomInt())
	builds := []string{testData["build_name0"], testData["build_name1"]}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCreateBuilds(t, builds, "")
		},
		CheckDestroy:      verifyDeleted(fqrn, testCheckWatch),
		ProviderFactories: testAccProviders(),
		Steps: []resource.TestStep{
			{
				Config: util.ExecuteTemplate(fqrn, multipleBuildsWatchTemplate, testData),
				Check:  verifyXrayWatch(fqrn, testData),
			},
		},
	})
}

func TestAccWatch_allBuilds(t *testing.T) {
	_, fqrn, resourceName := test.MkNames("watch-", "xray_watch")
	testData := util.MergeMaps(testDataWatch)

	testData["resource_name"] = resourceName
	testData["watch_name"] = fmt.Sprintf("xray-watch-%d", test.RandomInt())
	testData["policy_name_0"] = fmt.Sprintf("xray-policy-%d", test.RandomInt())
	testData["watch_type"] = "all-builds"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		CheckDestroy:      verifyDeleted(fqrn, testCheckWatch),
		ProviderFactories: testAccProviders(),
		Steps: []resource.TestStep{
			{
				Config: util.ExecuteTemplate(fqrn, allBuildsWatchTemplate, testData),
				Check: resource.ComposeTestCheckFunc(
					verifyXrayWatch(fqrn, testData),
					resource.TestCheckTypeSetElemAttr(fqrn, "watch_resource.*.ant_filter.*.exclude_patterns.*", "a*"),
					resource.TestCheckTypeSetElemAttr(fqrn, "watch_resource.*.ant_filter.*.exclude_patterns.*", "b*"),
					resource.TestCheckTypeSetElemAttr(fqrn, "watch_resource.*.ant_filter.*.include_patterns.*", "ab*"),
					resource.TestCheckTypeSetElemAttr(fqrn, "watch_resource.*.ant_filter.*.exclude_patterns.*", "c*"),
					resource.TestCheckTypeSetElemAttr(fqrn, "watch_resource.*.ant_filter.*.exclude_patterns.*", "d*"),
					resource.TestCheckTypeSetElemAttr(fqrn, "watch_resource.*.ant_filter.*.include_patterns.*", "cd*"),
				),
			},
		},
	})
}

func TestAccWatch_invalidBuildFilter(t *testing.T) {
	_, fqrn, resourceName := test.MkNames("watch-", "xray_watch")
	testData := util.MergeMaps(testDataWatch)

	testData["resource_name"] = resourceName
	testData["watch_name"] = fmt.Sprintf("xray-watch-%d", test.RandomInt())
	testData["policy_name_0"] = fmt.Sprintf("xray-policy-%d", test.RandomInt())

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		CheckDestroy:      verifyDeleted(fqrn, testCheckWatch),
		ProviderFactories: testAccProviders(),
		Steps: []resource.TestStep{
			{
				Config:      util.ExecuteTemplate(fqrn, invalidBuildsWatchFilterTemplate, testData),
				ExpectError: regexp.MustCompile(`attribute 'ant_filter' is set when 'watch_resource\.type' is not set to 'all-builds' or 'all-projects'`),
			},
		},
	})
}

func TestAccWatch_allProjects(t *testing.T) {
	_, fqrn, resourceName := test.MkNames("watch-", "xray_watch")
	testData := util.MergeMaps(testDataWatch)

	testData["resource_name"] = resourceName
	testData["watch_name"] = fmt.Sprintf("xray-watch-%d", test.RandomInt())
	testData["policy_name_0"] = fmt.Sprintf("xray-policy-%d", test.RandomInt())
	testData["watch_type"] = "all-projects"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		CheckDestroy:      verifyDeleted(fqrn, testCheckWatch),
		ProviderFactories: testAccProviders(),
		Steps: []resource.TestStep{
			{
				Config: util.ExecuteTemplate(fqrn, allProjectsWatchTemplate, testData),
				Check: resource.ComposeTestCheckFunc(
					verifyXrayWatch(fqrn, testData),
					resource.TestCheckTypeSetElemAttr(fqrn, "watch_resource.*.ant_filter.*.exclude_patterns.*", "a*"),
					resource.TestCheckTypeSetElemAttr(fqrn, "watch_resource.*.ant_filter.*.exclude_patterns.*", "b*"),
					resource.TestCheckTypeSetElemAttr(fqrn, "watch_resource.*.ant_filter.*.include_patterns.*", "ab*"),
				),
			},
		},
	})
}

//goland:noinspection GoErrorStringFormat
func TestAccWatch_singleProject(t *testing.T) {
	_, fqrn, resourceName := test.MkNames("watch-", "xray_watch")
	testData := util.MergeMaps(testDataWatch)

	testData["resource_name"] = resourceName
	testData["watch_name"] = fmt.Sprintf("xray-watch-%d", test.RandomInt())
	testData["policy_name_0"] = fmt.Sprintf("xray-policy-%d", test.RandomInt())
	testData["watch_type"] = "project"
	testData["project_name_0"] = fmt.Sprintf("test-project-%d", test.RandomInt())
	testData["project_name_1"] = fmt.Sprintf("test-project-%d", test.RandomInt())
	testData["project_key_0"] = "test1"
	testData["project_key_1"] = "test2"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCreateProject(t, testData["project_key_0"], testData["project_name_0"])
			testAccCreateProject(t, testData["project_key_1"], testData["project_name_1"])
		},
		CheckDestroy: verifyDeleted(fqrn, func(id string, request *resty.Request) (*resty.Response, error) {
			var errors []error
			if _, err := testAccDeleteProject(t, testData["project_key_0"]); err != nil {
				errors = append(errors, err)
			}
			if _, err := testAccDeleteProject(t, testData["project_key_1"]); err != nil {
				errors = append(errors, err)
			}
			if len(errors) > 0 {
				return nil, fmt.Errorf("errors during removing projects %s", errors)
			}
			//watch created by TF, so it will be automatically deleted by DeleteContext function
			testCheckPolicyDeleted(testData["policy_name_0"], t, request)
			return testCheckWatch(id, request)
		}),

		ProviderFactories: testAccProviders(),
		Steps: []resource.TestStep{
			{
				Config: util.ExecuteTemplate(fqrn, singleProjectWatchTemplate, testData),
				Check:  verifyXrayWatch(fqrn, testData),
			},
		},
	})
}

func TestAccWatch_invalidProjectFilter(t *testing.T) {
	_, fqrn, resourceName := test.MkNames("watch-", "xray_watch")
	testData := util.MergeMaps(testDataWatch)

	testData["resource_name"] = resourceName
	testData["watch_name"] = fmt.Sprintf("xray-watch-%d", test.RandomInt())
	testData["policy_name_0"] = fmt.Sprintf("xray-policy-%d", test.RandomInt())
	testData["watch_type"] = "project"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		CheckDestroy:      verifyDeleted(fqrn, testCheckWatch),
		ProviderFactories: testAccProviders(),
		Steps: []resource.TestStep{
			{

				Config:      util.ExecuteTemplate(fqrn, invalidProjectWatchFilterTemplate, testData),
				ExpectError: regexp.MustCompile(`attribute 'ant_filter' is set when 'watch_resource\.type' is not set to 'all-builds' or 'all-projects'`),
			},
		},
	})
}

const allReposSinglePolicyWatchTemplate = `resource "xray_security_policy" "security" {
  name        = "{{ .policy_name_0 }}"
  description = "Security policy description"
  type        = "security"
  rule {
    name     = "rule-name-severity"
    priority = 1
    criteria {
      min_severity = "High"
    }
    actions {
      webhooks = []
      mails    = ["test@email.com"]
      block_download {
        unscanned = true
        active    = true
      }
      block_release_bundle_distribution  = true
      fail_build                         = true
      notify_watch_recipients            = true
      notify_deployer                    = true
      create_ticket_enabled              = false
      build_failure_grace_period_in_days = 5
    }
  }
}

resource "xray_watch" "{{ .resource_name }}" {
  name        	= "{{ .watch_name }}"
  description 	= "{{ .description }}"
  active 		= {{ .active }}

  watch_resource {
	type       	= "{{ .watch_type }}"
	filter {
		type  	= "{{ .filter_type_0 }}"
		value	= "{{ .filter_value_0 }}"
	}
  }
  assigned_policy {
  	name 	= xray_security_policy.security.name
  	type 	= "security"
  }

  watch_recipients = ["{{ .watch_recipient_0 }}", "{{ .watch_recipient_1 }}"]
}`

const allReposMultiplePoliciesWatchTemplate = `resource "xray_security_policy" "security" {
  name        = "{{ .policy_name_0 }}"
  description = "Security policy description"
  type        = "security"
  rule {
    name     = "rule-name-severity"
    priority = 1
    criteria {
      min_severity = "High"
    }
    actions {
      webhooks = []
      mails    = ["test@email.com"]
      block_download {
        unscanned = true
        active    = true
      }
      block_release_bundle_distribution  = true
      fail_build                         = true
      notify_watch_recipients            = true
      notify_deployer                    = true
      create_ticket_enabled              = false
      build_failure_grace_period_in_days = 5
    }
  }
}

resource "xray_license_policy" "license" {
  name        = "{{ .policy_name_1 }}"
  description = "License policy description"
  type        = "license"
  rule {
    name     = "License_rule"
    priority = 1
    criteria {
      allowed_licenses         = ["Apache-1.0", "Apache-2.0"]
      allow_unknown            = false
      multi_license_permissive = true
    }
    actions {
      webhooks = []
      mails    = ["test@email.com"]
      block_download {
        unscanned = true
        active    = true
      }
      block_release_bundle_distribution  = false
      fail_build                         = true
      notify_watch_recipients            = true
      notify_deployer                    = true
      create_ticket_enabled              = false
      custom_severity                    = "High"
      build_failure_grace_period_in_days = 5
    }
  }
}

resource "xray_watch" "{{ .resource_name }}" {
  name        	= "{{ .watch_name }}"
  description 	= "{{ .description }}"
  active 		= {{ .active }}

  watch_resource {
	type       	= "{{ .watch_type }}"
	filter {
		type  	= "{{ .filter_type_0 }}"
		value	= "{{ .filter_value_0 }}"
	}
	filter {
		type  	= "{{ .filter_type_1 }}"
		value	= "{{ .filter_value_1 }}"
	}
  }
  assigned_policy {
  	name 	= xray_security_policy.security.name
  	type 	= "security"
  }
  assigned_policy {
  	name 	= xray_license_policy.license.name
  	type 	= "license"
  }

  watch_recipients = ["{{ .watch_recipient_0 }}", "{{ .watch_recipient_1 }}"]
}`

const singleRepositoryWatchTemplate = `resource "xray_security_policy" "security" {
  name        = "{{ .policy_name_0 }}"
  description = "Security policy description"
  type        = "security"
  rule {
    name     = "rule-name-severity"
    priority = 1
    criteria {
      min_severity = "High"
    }
    actions {
      webhooks = []
      mails    = ["test@email.com"]
      block_download {
        unscanned = true
        active    = true
      }
      block_release_bundle_distribution  = true
      fail_build                         = true
      notify_watch_recipients            = true
      notify_deployer                    = true
      create_ticket_enabled              = false
      build_failure_grace_period_in_days = 5
    }
  }
}

resource "xray_watch" "{{ .resource_name }}" {
  name        	= "{{ .watch_name }}"
  description 	= "{{ .description }}"
  active 		= {{ .active }}

  watch_resource {
	type       	= "{{ .watch_type }}"
	bin_mgr_id  = "default"
	name		= "{{ .repo0 }}"
	repo_type   = "{{ .repo_type }}"
	filter {
		type  	= "{{ .filter_type_0 }}"
		value	= "{{ .filter_value_0 }}"
	}
  }
  assigned_policy {
  	name 	= xray_security_policy.security.name
  	type 	= "security"
  }
  watch_recipients = ["{{ .watch_recipient_0 }}", "{{ .watch_recipient_1 }}"]
}`

const singleRepositoryInvalidWatchTemplate = `resource "xray_security_policy" "security" {
  name        = "{{ .policy_name_0 }}"
  description = "Security policy description"
  type        = "security"
  rule {
    name     = "rule-name-severity"
    priority = 1
    criteria {
      min_severity = "High"
    }
    actions {
      webhooks = []
      mails    = ["test@email.com"]
      block_download {
        unscanned = true
        active    = true
      }
      block_release_bundle_distribution  = true
      fail_build                         = true
      notify_watch_recipients            = true
      notify_deployer                    = true
      create_ticket_enabled              = false
      build_failure_grace_period_in_days = 5
    }
  }
}

resource "xray_watch" "{{ .resource_name }}" {
  name        	= "{{ .watch_name }}"
  description 	= "{{ .description }}"
  active 		= {{ .active }}

  watch_resource {
	type       	= "{{ .watch_type }}"
	bin_mgr_id  = "default"
	name		= "{{ .repo0 }}"
	filter {
		type  	= "{{ .filter_type_0 }}"
		value	= "{{ .filter_value_0 }}"
	}
  }
  assigned_policy {
  	name 	= xray_security_policy.security.name
  	type 	= "security"
  }
  watch_recipients = ["{{ .watch_recipient_0 }}", "{{ .watch_recipient_1 }}"]
}`

const multipleRepositoriesWatchTemplate = `resource "xray_security_policy" "security" {
  name        = "{{ .policy_name_0 }}"
  description = "Security policy description"
  type        = "security"
  rule {
    name     = "rule-name-severity"
    priority = 1
    criteria {
      min_severity = "High"
    }
    actions {
      webhooks = []
      mails    = ["test@email.com"]
      block_download {
        unscanned = true
        active    = true
      }
      block_release_bundle_distribution  = true
      fail_build                         = true
      notify_watch_recipients            = true
      notify_deployer                    = true
      create_ticket_enabled              = false
      build_failure_grace_period_in_days = 5
    }
  }
}

resource "xray_watch" "{{ .resource_name }}" {
  name        	= "{{ .watch_name }}"
  description 	= "{{ .description }}"
  active 		= {{ .active }}

  watch_resource {
	type       	= "{{ .watch_type }}"
	bin_mgr_id  = "default"
	name		= "{{ .repo0 }}"
	repo_type   = "{{ .repo_type }}"
	filter {
		type  	= "{{ .filter_type_0 }}"
		value	= "{{ .filter_value_0 }}"
	}
  }
  watch_resource {
	type       	= "repository"
	bin_mgr_id  = "default"
	name		= "{{ .repo1 }}"
	repo_type   = "{{ .repo_type }}"
	filter {
		type  	= "{{ .filter_type_0 }}"
		value	= "{{ .filter_value_0 }}"
	}
}
  assigned_policy {
  	name 	= xray_security_policy.security.name
  	type 	= "security"
}
  watch_recipients = ["{{ .watch_recipient_0 }}", "{{ .watch_recipient_1 }}"]
}`

const buildWatchTemplate = `resource "xray_security_policy" "security" {
  name        = "{{ .policy_name_0 }}"
  description = "Security policy description"
  type        = "security"
  rule {
    name     = "rule-name-severity"
    priority = 1
    criteria {
      min_severity = "High"
    }
    actions {
      webhooks = []
      mails    = ["test@email.com"]
      block_download {
        unscanned = true
        active    = true
      }
      block_release_bundle_distribution  = true
      fail_build                         = true
      notify_watch_recipients            = true
      notify_deployer                    = true
      create_ticket_enabled              = false
      build_failure_grace_period_in_days = 5
    }
  }
}

resource "xray_watch" "{{ .resource_name }}" {
  name        	= "{{ .watch_name }}"
  description 	= "{{ .description }}"
  active 		= {{ .active }}

  watch_resource {
	type       	= "{{ .watch_type }}"
	bin_mgr_id  = "default"
	name		= "{{ .build_name0 }}"
  }
  assigned_policy {
  	name 	= xray_security_policy.security.name
  	type 	= "security"
  }
  watch_recipients = ["{{ .watch_recipient_0 }}", "{{ .watch_recipient_1 }}"]
}`

const multipleBuildsWatchTemplate = `resource "xray_security_policy" "security" {
  name        = "{{ .policy_name_0 }}"
  description = "Security policy description"
  type        = "security"
  rule {
    name     = "rule-name-severity"
    priority = 1
    criteria {
      min_severity = "High"
    }
    actions {
      webhooks = []
      mails    = ["test@email.com"]
      block_download {
        unscanned = true
        active    = true
      }
      block_release_bundle_distribution  = true
      fail_build                         = true
      notify_watch_recipients            = true
      notify_deployer                    = true
      create_ticket_enabled              = false
      build_failure_grace_period_in_days = 5
    }
  }
}

resource "xray_watch" "{{ .resource_name }}" {
  name        	= "{{ .watch_name }}"
  description 	= "{{ .description }}"
  active 		= {{ .active }}

  watch_resource {
	type       	= "{{ .watch_type }}"
	bin_mgr_id  = "default"
	name		= "{{ .build_name0 }}"
  }
  watch_resource {
	type       	= "{{ .watch_type }}"
	bin_mgr_id  = "default"
	name		= "{{ .build_name1 }}"
  }
  assigned_policy {
  	name 	= xray_security_policy.security.name
  	type 	= "security"
  }
  watch_recipients = ["{{ .watch_recipient_0 }}", "{{ .watch_recipient_1 }}"]
}`

const allBuildsWatchTemplate = `resource "xray_security_policy" "security" {
  name        = "{{ .policy_name_0 }}"
  description = "Security policy description"
  type        = "security"
  rule {
    name     = "rule-name-severity"
    priority = 1
    criteria {
      min_severity = "High"
    }
    actions {
      webhooks = []
      mails    = ["test@email.com"]
      block_download {
        unscanned = true
        active    = true
      }
      block_release_bundle_distribution  = true
      fail_build                         = true
      notify_watch_recipients            = true
      notify_deployer                    = true
      create_ticket_enabled              = false
      build_failure_grace_period_in_days = 5
    }
  }
}

resource "xray_watch" "{{ .resource_name }}" {
  name        	= "{{ .watch_name }}"
  description 	= "{{ .description }}"
  active 		= {{ .active }}

  watch_resource {
	type       	= "{{ .watch_type }}"
	bin_mgr_id  = "default"
	ant_filter {
		exclude_patterns = ["a*", "b*"]
		include_patterns = ["ab*"]
	}
	ant_filter {
		exclude_patterns = ["c*", "d*"]
		include_patterns = ["cd*"]
	}
  }

  assigned_policy {
  	name 	= xray_security_policy.security.name
  	type 	= "security"
  }
  watch_recipients = ["{{ .watch_recipient_0 }}", "{{ .watch_recipient_1 }}"]
}`

const invalidBuildsWatchFilterTemplate = `resource "xray_security_policy" "security" {
  name        = "{{ .policy_name_0 }}"
  description = "Security policy description"
  type        = "security"
  rule {
    name     = "rule-name-severity"
    priority = 1
    criteria {
      min_severity = "High"
    }
    actions {
      webhooks = []
      mails    = ["test@email.com"]
      block_download {
        unscanned = true
        active    = true
      }
      block_release_bundle_distribution  = true
      fail_build                         = true
      notify_watch_recipients            = true
      notify_deployer                    = true
      create_ticket_enabled              = false
      build_failure_grace_period_in_days = 5
    }
  }
}

resource "xray_watch" "{{ .resource_name }}" {
  name        	= "{{ .watch_name }}"
  description 	= "{{ .description }}"
  active 		= {{ .active }}

  watch_resource {
	type       	= "build"
	bin_mgr_id  = "default"
	ant_filter {
		exclude_patterns = ["a*", "b*"]
		include_patterns = ["ab*"]
	}
  }

  assigned_policy {
  	name 	= xray_security_policy.security.name
  	type 	= "security"
  }
  watch_recipients = ["{{ .watch_recipient_0 }}", "{{ .watch_recipient_1 }}"]
}`

const allProjectsWatchTemplate = `resource "xray_security_policy" "security" {
  name        = "{{ .policy_name_0 }}"
  description = "Security policy description"
  type        = "security"
  rule {
    name     = "rule-name-severity"
    priority = 1
    criteria {
      min_severity = "High"
    }
    actions {
      webhooks = []
      mails    = ["test@email.com"]
      block_download {
        unscanned = true
        active    = true
      }
      block_release_bundle_distribution  = true
      fail_build                         = true
      notify_watch_recipients            = true
      notify_deployer                    = true
      create_ticket_enabled              = false
      build_failure_grace_period_in_days = 5
    }
  }
}

resource "xray_watch" "{{ .resource_name }}" {
  name        	= "{{ .watch_name }}"
  description 	= "{{ .description }}"
  active 		= {{ .active }}

  watch_resource {
	type       	= "all-projects"
	bin_mgr_id  = "default"
	ant_filter {
		exclude_patterns = ["a*", "b*"]
		include_patterns = ["ab*"]
	}
  }
  assigned_policy {
  	name 	= xray_security_policy.security.name
  	type 	= "security"
  }
  watch_recipients = ["{{ .watch_recipient_0 }}", "{{ .watch_recipient_1 }}"]
}`

const singleProjectWatchTemplate = `resource "xray_security_policy" "security" {
  name        = "{{ .policy_name_0 }}"
  description = "Security policy description"
  type        = "security"
  rule {
    name     = "rule-name-severity"
    priority = 1
    criteria {
      min_severity = "High"
    }
    actions {
      webhooks = []
      mails    = ["test@email.com"]
      block_download {
        unscanned = true
        active    = true
      }
      block_release_bundle_distribution  = true
      fail_build                         = true
      notify_watch_recipients            = true
      notify_deployer                    = true
      create_ticket_enabled              = false
      build_failure_grace_period_in_days = 5
    }
  }
}

resource "xray_watch" "{{ .resource_name }}" {
  name        	= "{{ .watch_name }}"
  description 	= "{{ .description }}"
  active 		= {{ .active }}

  watch_resource {
	type   = "project"
	name   = "{{ .project_key_0 }}"
  }
  watch_resource {
	type       	= "project"
	name 		= "{{ .project_key_1 }}"
  }
  assigned_policy {
  	name 	= xray_security_policy.security.name
  	type 	= "security"
  }
  watch_recipients = ["{{ .watch_recipient_0 }}", "{{ .watch_recipient_1 }}"]
}`

const invalidProjectWatchFilterTemplate = `resource "xray_security_policy" "security" {
  name        = "{{ .policy_name_0 }}"
  description = "Security policy description"
  type        = "security"
  rule {
    name     = "rule-name-severity"
    priority = 1
    criteria {
      min_severity = "High"
    }
    actions {
      webhooks = []
      mails    = ["test@email.com"]
      block_download {
        unscanned = true
        active    = true
      }
      block_release_bundle_distribution  = true
      fail_build                         = true
      notify_watch_recipients            = true
      notify_deployer                    = true
      create_ticket_enabled              = false
      build_failure_grace_period_in_days = 5
    }
  }
}

resource "xray_watch" "{{ .resource_name }}" {
  name        = "{{ .watch_name }}"
  description = "{{ .description }}"
  active      = {{ .active }}

  watch_resource {
	type = "project"
	name = "fake-project"
	ant_filter {
		exclude_patterns = ["a*"]
		include_patterns = ["b*"]
	}
  }

  assigned_policy {
  	name = xray_security_policy.security.name
  	type = "security"
  }
  watch_recipients = ["{{ .watch_recipient_0 }}", "{{ .watch_recipient_1 }}"]
}`

func verifyXrayWatch(fqrn string, testData map[string]string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(fqrn, "name", testData["watch_name"]),
		resource.TestCheckResourceAttr(fqrn, "description", testData["description"]),
		resource.TestCheckResourceAttr(fqrn, "watch_resource.0.type", testData["watch_type"]),
		resource.TestCheckResourceAttr(fqrn, "assigned_policy.0.name", testData["policy_name_0"]),
		resource.TestCheckResourceAttr(fqrn, "assigned_policy.0.type", "security"),
	)
}

func checkWatch(id string, request *resty.Request) (*resty.Response, error) {
	return request.Get("xray/api/v2/watches/" + id)
}

func testCheckWatch(id string, request *resty.Request) (*resty.Response, error) {
	return checkWatch(id, request.AddRetryCondition(client.NeverRetry))
}
