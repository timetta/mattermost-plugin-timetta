package main

import (
	"fmt"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
)

type Plugin struct {
	plugin.MattermostPlugin

	configuration configurationStore
	odata         *odataClient
}

func (p *Plugin) OnActivate() error {
	p.odata = newODataClient()
	return p.OnConfigurationChange()
}

func (p *Plugin) OnConfigurationChange() error {
	config := new(configuration)
	if err := p.API.LoadPluginConfiguration(config); err != nil {
		return fmt.Errorf("load plugin configuration: %w", err)
	}
	if err := config.prepare(); err != nil {
		return fmt.Errorf("invalid plugin configuration: %w", err)
	}
	p.configuration.set(config)
	return nil
}

func (p *Plugin) MessageWillBePosted(_ *plugin.Context, post *model.Post) (*model.Post, string) {
	p.enrichPost(post)
	return post, ""
}

func (p *Plugin) MessageWillBeUpdated(_ *plugin.Context, newPost, _ *model.Post) (*model.Post, string) {
	p.enrichPost(newPost)
	return newPost, ""
}

func (p *Plugin) enrichPost(post *model.Post) {
	if post == nil {
		return
	}
	config := p.configuration.get()
	if config == nil || p.odata == nil {
		return
	}
	links := findEntityLinks(post.Message, config)
	attachments := p.loadAttachments(config, links)
	post.Props = replaceTimettaAttachments(post.Props, attachments)
}
