package gh

import (
	"bytes"
	"errors"
	"gopkg.in/yaml.v3"
	"io"
	"log"
	"slices"
	"strings"
	"time"
)

const (
	GhURL = "https://github.com/"
)

type ConfigRepos []ConfigRepo

type ConfigRepo struct {
	Type  string       `yaml:"type"`
	Repos []Repository `yaml:"repo"`
	Qs    Qs           `yaml:"qs,omitempty"`
}

type Repository struct {
	LastUpdated time.Time
	Doc         string `yaml:"doc,omitempty"` // 该repo的官方文档URL
	URL         string `yaml:"url"`           // 该repo的gh URL
	Name        string `yaml:"name,omitempty"`
	User        string
	Des         string `yaml:"des,omitempty"`   // 描述
	Type        string `yaml:"type"`            // used to mark Type
	Alias       string `yaml:"alias,omitempty"` // 如果有alias，则直接渲染为[alias](URL)，而不是[User/Name](URL)
	Tag         string `yaml:"tag,omitempty"`   // 原本的文件名，比如说 db.yml, db.yml, ...
	Qs          Qs     `yaml:"qs,omitempty"`
	Cmd         Cmd    `yaml:"cmd,omitempty"`
	// Pix         []string `yaml:"pix"`
	// Use         []struct {
	// 	URL string `yaml:"url,omitempty"`
	// 	Des string `yaml:"des,omitempty"`
	// } `yaml:"use,omitempty"`
	Qq     `yaml:"qq,omitempty"`
	IsStar bool // 用来标识该repo是否在gh.yml中
}

type Qq []struct {
	Topic string `yaml:"topic"`
	URL   string `yaml:"url,omitempty"`
	Des   string `yaml:"des,omitempty"`
	Qs    Qs     `yaml:"qs,omitempty"`
}

type Qs []struct {
	Q string `yaml:"q,omitempty"` // 问题
	X string `yaml:"x,omitempty"` // 简要回答
	P string `yaml:"p,omitempty"`
	U string `yaml:"u,omitempty"` // url
}

type Cmd []struct {
	C string `yaml:"c"`
	X string `yaml:"x,omitempty"` // 该命令的描述
	K bool   `yaml:"k,omitempty"` // 该命令是否重要 default: false
}

type Repos []Repository

type Gh []string

func NewConfigRepos(f []byte) ConfigRepos {
	var ghs ConfigRepos

	d := yaml.NewDecoder(bytes.NewReader(f))
	for {
		// create new spec here
		spec := new(ConfigRepos)
		// pass a reference to spec reference
		if err := d.Decode(&spec); err != nil {
			// break the loop in case of EOF
			if errors.Is(err, io.EOF) {
				break
			}
			panic(err)
		}

		ghs = append(ghs, *spec...)
	}
	return ghs
}

// 给Repo的Tag字段赋值，否则为空字符串
func (cr ConfigRepos) WithTag(tag string) ConfigRepos {
	for _, repo := range cr {
		for i := range repo.Repos {
			repo.Repos[i].Tag = tag
		}
	}
	return cr
}

// MergeConfigs 将多个 ConfigRepos 合并为一个
// func MergeConfigs(in <-chan ConfigRepos) ConfigRepos {
// 	merged := make(ConfigRepos)
//
// 	for repo := range in {
// 		for k, v := range repo {
// 			merged[k] = v
// 		}
// 	}
//
// 	return merged
// }

// ToRepos Convert Type to Repo
func (cr ConfigRepos) ToRepos() Repos {
	repos := make(Repos, 0)
	for _, config := range cr {
		for _, repo := range config.Repos {
			if strings.Contains(repo.URL, GhURL) {
				sx, found := strings.CutPrefix(repo.URL, GhURL)
				// if !found {
				// 	log.Printf("Invalid URL: %s", repo.URL)
				// 	continue
				// }
				// splits := strings.Split(sx, "/")
				// if len(splits) != 2 {
				// 	log.Printf("URL Split Error: %s", repo.URL)
				// 	continue
				// }
				//
				// repo.User = splits[0]
				// repo.Name = splits[1]
				// repo.IsStar = true
				// repo.Type = config.Type
				// repos = append(repos, repo)

				if found {
					// splits := strings.Split(sx, "/")
					// if len(splits) == 2 {
					// 	repo.User = splits[0]
					// 	repo.Name = splits[1]
					// 	repo.IsStar = true
					// 	repo.Type = config.Type
					// 	repos = append(repos, repo)
					// } else if len(splits) > 2 {
					// 	// 确保 splits 不是 nil 并且有足够的元素
					// 	if splits[0] == "golang" && splits[1] == "go" {
					// 		curator := slices.Index(splits, "src")
					// 		if curator != -1 && curator < len(splits)-1 {
					// 			repo.User = splits[0]
					// 			repo.Name = splits[1] + "/" + strings.Join(splits[curator+1:], "/")
					// 			repo.IsStar = true
					// 			repo.Type = config.Type
					// 			repos = append(repos, repo)
					// 		} else {
					// 			log.Printf("Index Error: src not found in splits")
					// 		}
					// 	} else {
					// 		log.Printf("URL Split Error: not enough elements in splits")
					// 	}
					// } else {
					// 	log.Printf("URL Split Error: unexpected format")
					// }

					splits := strings.Split(sx, "/")
					if len(splits) == 2 {
						repo.User = splits[0]
						repo.Name = splits[1]
						repo.IsStar = true
						repo.Type = config.Type
						repos = append(repos, repo)
					} else {
						log.Printf("URL Split Error: unexpected format: %s", repo.URL)
					}
				} else {
					log.Printf("CutPrefix Error URL: %s", repo.URL)
				}
			} else {
				log.Printf("URL Not Contains: %s", repo.URL)
			}
		}
	}

	return repos
}

// ExtractTags Extract tags from Repos
func (rs Repos) ExtractTags() []string {
	var tags []string
	for _, repo := range rs {
		if repo.Type != "" && !slices.Contains(tags, repo.Type) {
			tags = append(tags, repo.Type)
		}
	}
	return tags
}

// QueryReposByTag Query Repos by Type
func (rs Repos) QueryReposByTag(tag string) Repos {
	var res Repos
	for _, repo := range rs {
		if repo.Type == tag {
			res = append(res, repo)
		}
	}
	return res
}

// FilterReposMD x
func (cr *ConfigRepos) FilterReposMD() ConfigRepos {
	var filteredConfig ConfigRepos
	for _, crv := range *cr {
		// if crv.Md {
		var filteredRepos []Repository
		for _, repo := range crv.Repos {
			if repo.Qs != nil {
				// repo.Pix = addMarkdownPicFormat(repo.Pix)
				filteredRepos = append(filteredRepos, repo)
			}
		}
		crv.Repos = filteredRepos
		filteredConfig = append(filteredConfig, crv)
		// }
	}
	return filteredConfig
}

// IsTypeQsEmpty 判断该type是否为空
func (cr *ConfigRepos) IsTypeQsEmpty() bool {
	for _, crv := range *cr {
		if crv.Qs == nil {
			return true
		}
	}
	return false
}

// func (cr *ConfigRepos) FilterWorksMD() ConfigRepos {
// 	var filteredConfig ConfigRepos
// 	for _, crv := range *cr {
// 		// if crv.Md {
// 		var filteredRepos []Repository
// 		crv.Repos = filteredRepos
// 		filteredConfig = append(filteredConfig, crv)
// 		// }
// 	}
// 	return filteredConfig
// }

// GetFileNameFromURL 从给定的 URL 中提取并返回文件名。
// func GetFileNameFromURL(urlString string) (string, error) {
// 	// 解析 URL
// 	parsedURL, err := url.Parse(urlString)
// 	if err != nil {
// 		return "", fmt.Errorf("error parsing URL: %v", err)
// 	}
//
// 	// 获取路径
// 	urlPath := parsedURL.Path
//
// 	// 获取文件名
// 	fileName := path.Base(urlPath)
//
// 	return fileName, nil
// }

// 用来渲染pic
// func addMarkdownPicFormat(URLs []string) []string {
// 	res := make([]string, len(URLs))
// 	for _, u := range URLs {
// 		name, _ := GetFileNameFromURL(u)
// 		res = append(res, fmt.Sprintf("![%s](%s)\n", name, u))
// 	}
// 	return res
// }
