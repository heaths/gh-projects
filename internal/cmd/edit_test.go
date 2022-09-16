package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/MakeNowJust/heredoc"
	"github.com/cli/go-gh/pkg/repository"
	"github.com/heaths/gh-projects/internal/utils"
	"github.com/heaths/go-console"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestNewEditCmd(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		stdin    *bytes.Buffer
		wantOpts *editOptions
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
			name: "only project number",
			args: []string{"1"},
			wantOpts: &editOptions{
				number: 1,
			},
		},
		{
			name: "basic parameters",
			args: []string{"1", "-t", "title", "-d", "description", "-b", "body", "--public"},
			wantOpts: &editOptions{
				number:      1,
				title:       "title",
				description: utils.Ptr("description"),
				body:        utils.Ptr("body"),
				public:      utils.Ptr(true),
			},
		},
		{
			name:  "body from stdin",
			args:  []string{"1", "-b", "-"},
			stdin: bytes.NewBufferString("stdin"),
			wantOpts: &editOptions{
				number: 1,
				body:   utils.Ptr("stdin"),
			},
		},
		{
			name:    "invalid issue number",
			args:    []string{"1", "--add-issue", "test"},
			wantErr: "invalid issue number: test",
		},
		{
			name: "issue number with hash prefix",
			args: []string{"1", "--add-issue", "#2"},
			wantOpts: &editOptions{
				number:    1,
				addIssues: []int{2},
			},
		},
		{
			name: "issue number",
			args: []string{"1", "--remove-issue", "2"},
			wantOpts: &editOptions{
				number:       1,
				removeIssues: []int{2},
			},
		},
		{
			name: "single field",
			args: []string{"1", "--add-issue", "2", "-f", "Status=Done"},
			wantOpts: &editOptions{
				number:    1,
				addIssues: []int{2},
				fields: map[string]string{
					"Status": "Done",
				},
			},
		},
		{
			name: "multiple field",
			args: []string{"1", "--add-issue", "2", "-f", "Status=Done,Iteration=Iteration 1", "-f", "Cost=1"},
			wantOpts: &editOptions{
				number:    1,
				addIssues: []int{2},
				fields: map[string]string{
					"Status":    "Done",
					"Iteration": "Iteration 1",
					"Cost":      "1",
				},
			},
		},
		{
			name:    "fields require issues",
			args:    []string{"1", "--field", "Status=Done"},
			wantErr: "--field requires --add-issue",
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

			var gotOpts *editOptions
			cmd := NewEditCmd(globalOpts, func(opts *editOptions) error {
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
			assert.Equal(t, tt.wantOpts.description, gotOpts.description)
			assert.Equal(t, tt.wantOpts.body, gotOpts.body)
			assert.Equal(t, tt.wantOpts.public, gotOpts.public)
			assert.Equal(t, tt.wantOpts.addIssues, gotOpts.addIssues)
			assert.Equal(t, tt.wantOpts.removeIssues, gotOpts.removeIssues)
			assert.Equal(t, tt.wantOpts.fields, gotOpts.fields)
		})
	}
}

func TestEdit(t *testing.T) {
	tests := []struct {
		name       string
		opts       *editOptions
		tty        bool
		mocks      func()
		wantStdout string
		wantErr    string
	}{
		{
			name: "change title",
			opts: &editOptions{
				number: 1,
				title:  "new title",
			},
			mocks: func() {
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(`{
						"data": {
							"repository": {
								"projectV2": {
									"id": "PN_1"
								}
							}
						}
					}`)
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(`{
						"data": {
							"updateProjectV2": {
								"projectV2": {
									"url": "https://github.com/users/heaths/projects/1"
								}
							}
						}
					}`)
			},
		},
		{
			name: "change title (tty)",
			opts: &editOptions{
				number: 1,
				title:  "new title",
			},
			tty: true,
			mocks: func() {
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(`{
						"data": {
							"repository": {
								"projectV2": {
									"id": "PN_1"
								}
							}
						}
					}`)
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(`{
						"data": {
							"updateProjectV2": {
								"projectV2": {
									"url": "https://github.com/users/heaths/projects/1"
								}
							}
						}
					}`)
			},
			wantStdout: "https://github.com/users/heaths/projects/1\n",
		},
		{
			name: "add issues with fields (tty)",
			opts: &editOptions{
				GlobalOptions: GlobalOptions{
					Verbose: true,
				},
				number:    1,
				addIssues: []int{2, 3},
				fields: map[string]string{
					"status":    "todo",
					"iteration": "iteration 1",
				},
			},
			tty: true,
			mocks: func() {
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(`{
						"data": {
							"repository": {
								"projectV2": {
									"id": "PN_1"
								}
							}
						}
					}`)
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(`{
						"data": {
							"updateProjectV2": {
								"projectV2": {
									"url": "https://github.com/users/heaths/projects/1"
								}
							}
						}
					}`)
				// getFields also testing paging and short-circuiting subsequent page fetches
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(`{
						"data": {
							"repository": {
								"projectV2": {
									"fields": {
										"nodes": [
											{
												"id": "PNF_Title",
												"name": "Title",
												"dataType": "TITLE"
											}
										],
										"pageInfo": {
											"hasNextPage": true,
											"endCursor": "PAGE2"
										}
									}
								}
							}
						}
					}`)
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(`{
						"data": {
							"repository": {
								"projectV2": {
									"fields": {
										"nodes": [
											{
												"id": "PNF_Status",
												"name": "Status",
												"dataType": "SINGLE_SELECT",
												"options": [
													{
														"id": "PNF_Status_Todo",
														"name": "Todo"
													},
													{
														"id": "PNF_Status_InProgress",
														"name": "In Progress"
													},
													{
														"id": "PNF_Status_Done",
														"name": "Done"
													}
												]
											},
											{
												"id": "PNF_Labels",
												"name": "Labels",
												"dataType": "LABELS"
											},
											{
												"id": "PNF_Iteration",
												"name": "Iteration",
												"dataType": "ITERATION",
												"configuration": {
													"iterations": [
														{
															"id": "PNF_Iteration_1",
															"name": "Iteration 1"
														},
														{
															"id": "PNF_Iteration_2",
															"name": "Iteration 2"
														}
													]
												}
											}
										],
										"pageInfo": {
											"hasNextPage": true,
											"endCursor": "PAGE3"
										}
									}
								}
							}
						}
					}`)
				// addIssue 1
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(`{
						"data": {
							"repository": {
								"issueOrPullRequest": {
									"id": "I_2"
								}
							}
						}
					}`)
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(`{
						"data": {
							"addProjectV2ItemById": {
								"item": {
									"id": "PNI_2"
								}
							}
						}
					}`)
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(`{
						"data": {
							"updateProjectV2ItemFieldValue": {
								"projectV2Item": {
									"id": "PNI_1"
								}
							}
						}
					}`)
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(`{
						"data": {
							"updateProjectV2ItemFieldValue": {
								"projectV2Item": {
									"id": "PNI_1"
								}
							}
						}
					}`)
				// addIssue 2
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(`{
						"data": {
							"repository": {
								"issueOrPullRequest": {
									"id": "I_2"
								}
							}
						}
					}`)
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(`{
						"data": {
							"addProjectV2ItemById": {
								"item": {
									"id": "PNI_2"
								}
							}
						}
					}`)
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(`{
						"data": {
							"updateProjectV2ItemFieldValue": {
								"projectV2Item": {
									"id": "PNI_1"
								}
							}
						}
					}`)
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(`{
						"data": {
							"updateProjectV2ItemFieldValue": {
								"projectV2Item": {
									"id": "PNI_1"
								}
							}
						}
					}`)
			},
			wantStdout: heredoc.Doc(`
				Added 2 issues
				https://github.com/users/heaths/projects/1
			`),
		},
		{
			name: "undefined field",
			opts: &editOptions{
				number:    1,
				addIssues: []int{2},
				fields:    map[string]string{"Undefined": "true"},
			},
			tty: true,
			mocks: func() {
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(`{
						"data": {
							"repository": {
								"projectV2": {
									"id": "PN_1"
								}
							}
						}
					}`)
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(`{
						"data": {
							"updateProjectV2": {
								"projectV2": {
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
							"repository": {
								"projectV2": {
									"fields": {
										"nodes": [
											{
												"id": "PNF_Title",
												"name": "Title",
												"dataType": "TITLE"
											}
										],
										"pageInfo": {
											"hasNextPage": true,
											"endCursor": "PAGE2"
										}
									}
								}
							}
						}
					}`)
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(`{
						"data": {
							"repository": {
								"projectV2": {
									"fields": {
										"nodes": [
											{
												"id": "PNF_Status",
												"name": "Status",
												"dataType": "SINGLE_SELECT",
												"options": [
													{
														"id": "PNF_Status_Todo",
														"name": "Todo"
													},
													{
														"id": "PNF_Status_InProgress",
														"name": "In Progress"
													},
													{
														"id": "PNF_Status_Done",
														"name": "Done"
													}
												]
											},
											{
												"id": "PNF_Labels",
												"name": "Labels",
												"dataType": "LABELS"
											}
										],
										"pageInfo": {
											"hasNextPage": true,
											"endCursor": "PAGE3"
										}
									}
								}
							}
						}
					}`)
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(`{
						"data": {
							"repository": {
								"projectV2": {
									"fields": {
										"nodes": [
											{
												"id": "PNF_Cost",
												"name": "Cost",
												"dataType": "NUMBER"
											}
										],
										"pageInfo": {
											"hasNextPage": false,
											"endCursor": null
										}
									}
								}
							}
						}
					}`)
			},
			wantErr: `field "Undefined" not defined`,
		},
		{
			name: "issue not found",
			opts: &editOptions{
				number:    1,
				addIssues: []int{99},
				fields:    map[string]string{"status": "todo"},
			},
			mocks: func() {
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(`{
						"data": {
							"repository": {
								"projectV2": {
									"id": "PN_1"
								}
							}
						}
					}`)
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(`{
						"data": {
							"updateProjectV2": {
								"projectV2": {
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
							"repository": {
								"projectV2": {
									"fields": {
										"nodes": [
											{
												"id": "PNF_Status",
												"name": "Status",
												"dataType": "SINGLE_SELECT",
												"options": [
													{
														"id": "PNF_Status_Todo",
														"name": "Todo"
													},
													{
														"id": "PNF_Status_InProgress",
														"name": "In Progress"
													},
													{
														"id": "PNF_Status_Done",
														"name": "Done"
													}
												]
											}
										],
										"pageInfo": {
											"hasNextPage": true,
											"endCursor": "PAGE2"
										}
									}
								}
							}
						}
					}`)
				// addIssue 99
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(`{
						"data": {
							"repository": {
								"issueOrPullRequest": null
							}
						},
						"errors": [
							{
								"type": "NOT_FOUND",
								"message": "Could not resolve to an issue or pull request with the number of 99."
							}
						]
					}`)
			},
			wantErr: "GraphQL: Could not resolve to an issue or pull request with the number of 99.",
		},
		{
			name: "invalid field value",
			opts: &editOptions{
				number:    1,
				addIssues: []int{2},
				fields:    map[string]string{"Cost": "Huge"},
			},
			tty: true,
			mocks: func() {
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(`{
						"data": {
							"repository": {
								"projectV2": {
									"id": "PN_1"
								}
							}
						}
					}`)
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(`{
						"data": {
							"updateProjectV2": {
								"projectV2": {
									"url": "https://github.com/users/heaths/projects/1"
								}
							}
						}
					}`)
				// getFields
				gock.New("https://api.github.com").
					Post("/graphql").
					Reply(200).
					JSON(`{
						"data": {
							"repository": {
								"projectV2": {
									"fields": {
										"nodes": [
											{
												"id": "PNF_Cost",
												"name": "Cost",
												"dataType": "NUMBER"
											}
										],
										"pageInfo": {
											"hasNextPage": false,
											"endCursor": null
										}
									}
								}
							}
						}
					}`)
			},
			wantErr: `invalid number for field "Cost": Huge`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(gock.Off)

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
				tt.opts = &editOptions{}
			}
			tt.opts.GlobalOptions = *globalOpts
			tt.opts.workerCount = 1

			if tt.mocks != nil {
				tt.mocks()
			}

			err = edit(tt.opts)
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

func pendingMocks(mocks []gock.Mock) string {
	paths := make([]string, len(mocks))
	for i, mock := range mocks {
		paths[i] = mock.Request().URLStruct.String()
	}

	return fmt.Sprintf("%d unmatched mocks: %s", len(paths), strings.Join(paths, ", "))
}
