package slack

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type UserListMember struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Deleted  bool   `json:"deleted"`
	RealName string `json:"real_name"`
	Profile  struct {
		Email string `json:"email"`
	} `json:"profile"`
}

type UserListResponse struct {
	Ok      bool             `json:"ok"`
	Error   string           `json:"error"`
	Members []UserListMember `json:"members"`
}

type SlackInvite struct {
	Email    string `json:"email"`
	RealName string `json:"real_name"`
}

func resourceUser() *schema.Resource {
	return &schema.Resource{
		Create: resourceUserCreate,
		Read:   resourceUserRead,
		// Update is optional
		Update: resourceUserUpdate,
		Delete: resourceUserDelete,

		Schema: map[string]*schema.Schema{
			"email": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"full_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func findSlackMemberByAttribute(config *Config, eqAttributesFun func(userListMember UserListMember) bool) *UserListMember {
	resp, err := http.Get("https://slack.com/api/users.list?token=" + config.Token)

	if err != nil {
		log.Println("error GETing httpbin.org", err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	var userListResponse UserListResponse
	json.Unmarshal([]byte(body), &userListResponse)

	if !userListResponse.Ok {
		log.Println("Request not Ok: ", userListResponse)
	}

	for _, member := range userListResponse.Members {
		if eqAttributesFun(member) {
			return &member
		}
	}

	return nil
}

func resourceUserCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	email := d.Get("email").(string)
	fullName := d.Get("full_name").(string)

	d.SetId("")

	var slackInvite = SlackInvite{
		Email:    d.Get("email").(string),
		RealName: d.Get("full_name").(string),
	}

	slackInviteBytes, err := json.Marshal(slackInvite)

	if err != nil {
		return err
	}

	client := &http.Client{}
	req, _ := http.NewRequest("POST", "https://slack.com/api/users.admin.invite?email="+email+"&full_name="+url.QueryEscape(fullName), bytes.NewBuffer(slackInviteBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.Token)
	res, errRsp := client.Do(req)

	if errRsp != nil {
		return nil
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return nil
	}

	var inviteResponse UserListResponse
	json.Unmarshal([]byte(body), &inviteResponse)

	if !inviteResponse.Ok {
		log.Println("Error while trying invite new user: '", slackInvite.Email, "' to slack: '", inviteResponse.Error, "'")
		return nil
	}

	var slackMember = findSlackMemberByAttribute(config, func(userListMember UserListMember) bool {
		return userListMember.Profile.Email == slackInvite.Email
	})

	if slackMember != nil {
		d.SetId(slackMember.Id)
		// The create and update function should always return the read function to ensure the state is reflected in the terraform.state file
		return resourceUserRead(d, meta)
	} else {
		return nil
	}
}

func resourceUserRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	var slackMember = findSlackMemberByAttribute(config, func(userListMember UserListMember) bool {
		return userListMember.Id == d.Id()
	})

	if slackMember == nil {
		log.Println("Didn't found slackMember with id: ", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("email", slackMember.Profile.Email)
	d.Set("full_name", slackMember.RealName)

	return nil
}

func resourceUserUpdate(d *schema.ResourceData, meta interface{}) error {
	// The create and update function should always return the read function to ensure the state is reflected in the terraform.state file
	return resourceUserRead(d, meta)
}

func resourceUserDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}
