---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "xray_watch Resource - terraform-provider-xray"
subcategory: "Watch"
---

Provides an Xray Watch resource.

[Official documentation](https://www.jfrog.com/confluence/display/JFROG/Configuring+Xray+Watches#ConfiguringXrayWatches-CreatingaWatch).

[API documentation](https://www.jfrog.com/confluence/display/JFROG/Xray+REST+API#XrayRESTAPI-CreateWatch).


## Example Usage

```terraform
resource "xray_watch" "all-repos" {
  name        = "all-repos-watch"
  description = "Watch for all repositories, matching the filter"
  active      = true
  project_key = "testproj"

  watch_resource {
    type = "all-repos"

    filter {
      type  = "regex"
      value = ".*"
    }
  }

  assigned_policy {
    name = xray_security_policy.allowed_licenses.name
    type = "security"
  }

  assigned_policy {
    name = xray_license_policy.banned_licenses.name
    type = "license"
  }

  watch_recipients = ["test@email.com", "test1@email.com"]
}

resource "xray_watch" "repository" {
  name        = "repository-watch"
  description = "Watch a single repo or a list of repositories"
  active      = true
  project_key = "testproj"

  watch_resource {
    type       = "repository"
    bin_mgr_id = "default"
    name       = "your-repository-name"
    repo_type  = "local"

    filter {
      type  = "regex"
      value = ".*"
    }
  }

  watch_resource {
    type       = "repository"
    bin_mgr_id = "default"
    name       = "your-other-repository-name"
    repo_type  = "remote"

    filter {
      type  = "regex"
      value = ".*"
    }
  }

  assigned_policy {
    name = xray_security_policy.min_severity.name
    type = "security"
  }

  assigned_policy {
    name = xray_license_policy.cvss_range.name
    type = "license"
  }

  watch_recipients = ["test@email.com", "test1@email.com"]
}

resource "xray_watch" "all-builds-with-filters" {
  name        = "build-watch"
  description = "Watch all builds with Ant patterns filter"
  active      = true
  project_key = "testproj"

  watch_resource {
    type       = "all-builds"
    bin_mgr_id = "default"

    ant_filter {
      exclude_patterns = ["a*", "b*"]
      include_patterns = ["ab*"]
    }
  }

  assigned_policy {
    name = xray_security_policy.min_severity.name
    type = "security"
  }

  assigned_policy {
    name = xray_license_policy.cvss_range.name
    type = "license"
  }

  watch_recipients = ["test@email.com", "test1@email.com"]
}

resource "xray_watch" "build" {
  name        = "build-watch"
  description = "Watch a single build or a list of builds"
  active      = true
  project_key = "testproj"

  watch_resource {
    type       = "build"
    bin_mgr_id = "default"
    name       = "your-build-name"
  }

  watch_resource {
    type       = "build"
    bin_mgr_id = "default"
    name       = "your-other-build-name"
  }

  assigned_policy {
    name = xray_security_policy.min_severity.name
    type = "security"
  }

  assigned_policy {
    name = xray_license_policy.cvss_range.name
    type = "license"
  }

  watch_recipients = ["test@email.com", "test1@email.com"]
}

resource "xray_watch" "all-projects" {
  name        = "projects-watch"
  description = "Watch all the projects"
  active      = true
  project_key = "testproj"

  watch_resource {
    type       = "all-projects"
    bin_mgr_id = "default"
  }

  assigned_policy {
    name = xray_security_policy.min_severity.name
    type = "security"
  }

  assigned_policy {
    name = xray_license_policy.cvss_range.name
    type = "license"
  }

  watch_recipients = ["test@email.com", "test1@email.com"]
}

resource "xray_watch" "all-projects-with-filters" {
  name        = "projects-watch"
  description = "Watch all the projects with Ant patterns filter"
  active      = true
  project_key = "testproj"

  watch_resource {
    type       = "all-projects"
    bin_mgr_id = "default"

    ant_filter {
      exclude_patterns = ["a*", "b*"]
      include_patterns = ["ab*"]
    }
  }

  assigned_policy {
    name = xray_security_policy.min_severity.name
    type = "security"
  }

  assigned_policy {
    name = xray_license_policy.cvss_range.name
    type = "license"
  }

  watch_recipients = ["test@email.com", "test1@email.com"]
}

resource "xray_watch" "project" {
  name        = "project-watch"
  description = "Watch selected projects"
  active      = true
  project_key = "testproj"

  watch_resource {
    type = "project"
    name = "test"
  }
  watch_resource {
    type = "project"
    name = "test1"
  }

  assigned_policy {
    name = xray_security_policy.min_severity.name
    type = "security"
  }

  assigned_policy {
    name = xray_license_policy.cvss_range.name
    type = "license"
  }

  watch_recipients = ["test@email.com", "test1@email.com"]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `assigned_policy` (Block Set, Min: 1) Nested argument describing policies that will be applied. Defined below. (see [below for nested schema](#nestedblock--assigned_policy))
- `name` (String) Name of the watch (must be unique)
- `watch_resource` (Block Set, Min: 1) Nested argument describing the resources to be watched. Defined below. (see [below for nested schema](#nestedblock--watch_resource))

### Optional

- `active` (Boolean) Whether or not the watch is active
- `description` (String) Description of the watch
- `project_key` (String) Project key for assigning this resource to. Must be 3 - 10 lowercase alphanumeric and hyphen characters. Support repository and build watch resource types. When specifying individual repository or build they must be already assigned to the project. Build must be added as indexed resources.
- `watch_recipients` (Set of String) A list of email addressed that will get emailed when a violation is triggered.

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedblock--assigned_policy"></a>
### Nested Schema for `assigned_policy`

Required:

- `name` (String) The name of the policy that will be applied
- `type` (String) The type of the policy - security or license


<a id="nestedblock--watch_resource"></a>
### Nested Schema for `watch_resource`

Required:

- `type` (String) Type of resource to be watched. Options: `all-repos`, `repository`, `all-builds`, `build`, `project`, `all-projects`.

Optional:

- `ant_filter` (Block Set) `ant-patterns` filter for `all-builds` and `all-projects` watch_resource.type (see [below for nested schema](#nestedblock--watch_resource--ant_filter))
- `bin_mgr_id` (String) The ID number of a binary manager resource. Default value is `default`. To check the list of available binary managers, use the API call `${JFROG_URL}/xray/api/v1/binMgr` as an admin user, use `binMgrId` value. More info [here](https://www.jfrog.com/confluence/display/JFROG/Xray+REST+API#XrayRESTAPI-GetBinaryManager)
- `filter` (Block Set) Filter for `regex` and `package-type` type. Works only with `all-repos` watch_resource.type. (see [below for nested schema](#nestedblock--watch_resource--filter))
- `name` (String) The name of the build, repository or project. Xray indexing must be enabled on the repository or build
- `repo_type` (String) Type of repository. Only applicable when `type` is `repository`. Options: `local` or `remote`.

<a id="nestedblock--watch_resource--ant_filter"></a>
### Nested Schema for `watch_resource.ant_filter`

Required:

- `exclude_patterns` (List of String) List of Ant patterns.
- `include_patterns` (List of String) List of Ant patterns.


<a id="nestedblock--watch_resource--filter"></a>
### Nested Schema for `watch_resource.filter`

Required:

- `type` (String) The type of filter, such as `regex` or `package-type`
- `value` (String) The value of the filter, such as the text of the regex or name of the package type.