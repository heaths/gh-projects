package cmd

import (
	"bytes"
	"testing"

	"github.com/cli/go-gh/pkg/repository"
	"github.com/heaths/gh-projects/internal/utils"
	"github.com/heaths/go-console"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestNewCloneCmd(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		stdin    *bytes.Buffer
		wantOpts *cloneOptions
		wantErr  string
	}{
		{
			name:    "no args",
			wantErr: "missing required project number",
		},
		{
			name:    "invalid project number",
			args:    []string{"test"},
			wantErr: "invalid project number: test",
		},
		{
			name:    "only project number",
			args:    []string{"1"},
			wantErr: `required flag(s) "title" not set`,
		},
		{
			name: "basic parameters",
			args: []string{"1", "-t", "title", "-d", "description", "-b", "body", "--public"},
			wantOpts: &cloneOptions{
				projectOptions: projectOptions{
					number:      1,
					title:       "title",
					description: utils.Ptr("description"),
					body:        utils.Ptr("body"),
					public:      utils.Ptr(true),
				},
			},
		},
		{
			name:  "body from stdin",
			args:  []string{"1", "-t", "title", "-b", "-"},
			stdin: bytes.NewBufferString("stdin"),
			wantOpts: &cloneOptions{
				projectOptions: projectOptions{
					number: 1,
					title:  "title",
					body:   utils.Ptr("stdin"),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := console.Fake()
			if tt.stdin != nil {
				_, _, stdin := fake.Buffers()
				*stdin = *tt.stdin
			}

			globalOpts := &GlobalOptions{
				Console: fake,
			}

			var gotOpts *cloneOptions
			cmd := NewCloneCmd(globalOpts, func(opts *cloneOptions) error {
				gotOpts = opts
				return nil
			})
			cmd.SilenceUsage = true

			cmd.SetArgs(tt.args)
			err := cmd.Execute()

			if tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantOpts.number, gotOpts.number)
			assert.Equal(t, tt.wantOpts.title, gotOpts.title)
		})
	}
}

func TestClone(t *testing.T) {
	tests := []struct {
		name       string
		opts       *cloneOptions
		tty        bool
		mocks      func()
		wantStdout string
		wantErr    string
	}{
		{
			name: "clone",
			opts: &cloneOptions{
				projectOptions: projectOptions{
					number: 1,
					title:  "title",
				},
			},
			mocks: func() {
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(`{
						"data": {
							"viewer": {
								"id": "U_1"
							},
							"repository": {
								"projectV2": {
									"id": "PN_1",
									"url": "https://github.com/users/heaths/projects/1"
								}
							}
						}
					}`)
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(`{
						"data": {
							"copyProjectV2": {
								"projectV2": {
									"id": "PN_2",
									"url": "https://github.com/users/heaths/projects/2"
								}
							}
						}
					}`)
			},
		},
		{
			name: "clone (tty)",
			opts: &cloneOptions{
				projectOptions: projectOptions{
					number: 1,
					title:  "title",
				},
			},
			tty: true,
			mocks: func() {
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(`{
						"data": {
							"viewer": {
								"id": "U_1"
							},
							"repository": {
								"projectV2": {
									"id": "PN_1",
									"url": "https://github.com/users/heaths/projects/1"
								}
							}
						}
					}`)
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(`{
						"data": {
							"copyProjectV2": {
								"projectV2": {
									"id": "PN_2",
									"url": "https://github.com/users/heaths/projects/2"
								}
							}
						}
					}`)
			},
			wantStdout: "https://github.com/users/heaths/projects/2\n",
		},
		{
			name: "project not found",
			opts: &cloneOptions{
				projectOptions: projectOptions{
					number: 99,
					title:  "title",
				},
			},
			mocks: func() {
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(`{
						"data": {
							"viewer": {
								"id": "U_1"
							},
							"repository": {
								"projectV2": null
							}
						},
						"errors": [
							{
								"type": "NOT_FOUND",
								"message": "Could not resolve to a ProjectV2 with the number 99."
							}
						]
					}`)
			},
			wantErr: "GraphQL: Could not resolve to a ProjectV2 with the number 99.",
		},
		{
			name: "clone with parameters",
			opts: &cloneOptions{
				projectOptions: projectOptions{
					number:      1,
					title:       "title",
					description: utils.Ptr("description"),
					body:        utils.Ptr("body"),
					public:      utils.Ptr(true),
				},
			},
			mocks: func() {
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(`{
						"data": {
							"viewer": {
								"id": "U_1"
							},
							"repository": {
								"projectV2": {
									"id": "PN_1",
									"url": "https://github.com/users/heaths/projects/1",
									"public": false
								}
							}
						}
					}`)
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(`{
						"data": {
							"copyProjectV2": {
								"projectV2": {
									"id": "PN_2",
									"url": "https://github.com/users/heaths/projects/2"
								}
							}
						}
					}`)
				gock.New("https://api.github.com").
					Post("/graphql").
					MatchHeader("Content-Type", "application/json; charset=utf-8").
					JSON(map[string]interface{}{
						"query": mutationUpdateProjectV2,
						"variables": map[string]interface{}{
							"id":          "PN_2",
							"title":       "title",
							"description": "description",
							"body":        "body",
							"public":      true,
						},
					}).
					Reply(200).
					JSON(`{
						"data": {
							"updateProjectV2": {
								"projectV2": {
									"url": "https://github.com/users/heaths/projects/2"
								}
							}
						}
					}`)
			},
		},
		{
			name: "copies public visibility",
			opts: &cloneOptions{
				projectOptions: projectOptions{
					number: 1,
					title:  "title",
				},
			},
			mocks: func() {
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(`{
						"data": {
							"viewer": {
								"id": "U_1"
							},
							"repository": {
								"projectV2": {
									"id": "PN_1",
									"url": "https://github.com/users/heaths/projects/1",
									"public": true
								}
							}
						}
					}`)
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(`{
						"data": {
							"copyProjectV2": {
								"projectV2": {
									"id": "PN_2",
									"url": "https://github.com/users/heaths/projects/2"
								}
							}
						}
					}`)
				gock.New("https://api.github.com").
					Post("/graphql").
					MatchHeader("Content-Type", "application/json; charset=utf-8").
					JSON(map[string]interface{}{
						"query": mutationUpdateProjectV2,
						"variables": map[string]interface{}{
							"id":     "PN_2",
							"title":  "title",
							"public": true,
						},
					}).
					Reply(200).
					JSON(`{
						"data": {
							"updateProjectV2": {
								"projectV2": {
									"url": "https://github.com/users/heaths/projects/2"
								}
							}
						}
					}`)
			},
		},
		{
			name: "overrides private visibility",
			opts: &cloneOptions{
				projectOptions: projectOptions{
					number: 1,
					title:  "title",
					public: utils.Ptr(true),
				},
			},
			mocks: func() {
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(`{
						"data": {
							"viewer": {
								"id": "U_1"
							},
							"repository": {
								"projectV2": {
									"id": "PN_1",
									"url": "https://github.com/users/heaths/projects/1",
									"public": false
								}
							}
						}
					}`)
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(`{
						"data": {
							"copyProjectV2": {
								"projectV2": {
									"id": "PN_2",
									"url": "https://github.com/users/heaths/projects/2"
								}
							}
						}
					}`)
				gock.New("https://api.github.com").
					Post("/graphql").
					MatchHeader("Content-Type", "application/json; charset=utf-8").
					JSON(map[string]interface{}{
						"query": mutationUpdateProjectV2,
						"variables": map[string]interface{}{
							"id":     "PN_2",
							"title":  "title",
							"public": true,
						},
					}).
					Reply(200).
					JSON(`{
						"data": {
							"updateProjectV2": {
								"projectV2": {
									"url": "https://github.com/users/heaths/projects/2"
								}
							}
						}
					}`)
			},
		},
		{
			name: "overrides public visibility",
			opts: &cloneOptions{
				projectOptions: projectOptions{
					number: 1,
					title:  "title",
					public: utils.Ptr(false),
				},
			},
			mocks: func() {
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(`{
						"data": {
							"viewer": {
								"id": "U_1"
							},
							"repository": {
								"projectV2": {
									"id": "PN_1",
									"url": "https://github.com/users/heaths/projects/1",
									"public": true
								}
							}
						}
					}`)
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(`{
						"data": {
							"copyProjectV2": {
								"projectV2": {
									"id": "PN_2",
									"url": "https://github.com/users/heaths/projects/2"
								}
							}
						}
					}`)
				gock.New("https://api.github.com").
					Post("/graphql").
					MatchHeader("Content-Type", "application/json; charset=utf-8").
					JSON(map[string]interface{}{
						"query": mutationUpdateProjectV2,
						"variables": map[string]interface{}{
							"id":     "PN_2",
							"title":  "title",
							"public": false,
						},
					}).
					Reply(200).
					JSON(`{
						"data": {
							"updateProjectV2": {
								"projectV2": {
									"url": "https://github.com/users/heaths/projects/2"
								}
							}
						}
					}`)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(gock.Off)
			// gock.Observe(gock.DumpRequest)

			fake := console.Fake(console.WithStdoutTTY(tt.tty))
			repo, err := repository.Parse("heaths/gh-projects")
			assert.NoError(t, err)

			globalOpts := &GlobalOptions{
				Console: fake,
				Repo:    repo,
				Verbose: tt.opts.Verbose,

				authToken: "***",
				host:      "github.com",
			}

			if tt.opts == nil {
				tt.opts = &cloneOptions{}
			}
			tt.opts.GlobalOptions = *globalOpts

			if tt.mocks != nil {
				tt.mocks()
			}

			err = clone(tt.opts)
			if tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
				return
			}

			assert.NoError(t, err)
			assert.True(t, gock.IsDone(), pendingMocks(gock.Pending()))

			stdout, _, _ := fake.Buffers()
			assert.Equal(t, tt.wantStdout, stdout.String())
		})
	}
}
