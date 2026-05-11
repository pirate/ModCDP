package modcdp

import (
	"github.com/pirate/ModCDP/go/modcdp/client"
	"github.com/pirate/ModCDP/go/modcdp/injector"
	"github.com/pirate/ModCDP/go/modcdp/launcher"
	"github.com/pirate/ModCDP/go/modcdp/transport"
)

type ModCDPClient = client.ModCDPClient
type Options = client.Options
type LauncherConfig = client.LauncherConfig
type UpstreamConfig = client.UpstreamConfig
type InjectorConfig = client.InjectorConfig
type ClientConfig = client.ClientConfig
type ServerConfig = client.ServerConfig
type CustomCommand = client.CustomCommand
type CustomEvent = client.CustomEvent
type CustomMiddleware = client.CustomMiddleware
type LaunchOptions = launcher.LaunchOptions
type LocalBrowserLauncher = launcher.LocalBrowserLauncher
type RemoteBrowserLauncher = launcher.RemoteBrowserLauncher
type BrowserbaseBrowserLauncher = launcher.BrowserbaseBrowserLauncher
type NoopBrowserLauncher = launcher.NoopBrowserLauncher
type ExtensionInjector = injector.ExtensionInjector
type UpstreamTransport = transport.UpstreamTransport
type TargetTargetCreatedEvent = client.TargetTargetCreatedEvent
type TargetSetDiscoverTargetsParams = client.TargetSetDiscoverTargetsParams
type TargetCreateTargetParams = client.TargetCreateTargetParams
type TargetActivateTargetParams = client.TargetActivateTargetParams
type TargetTargetID = client.TargetTargetID

var New = client.New
var Bool = client.Bool
var NewLocalBrowserLauncher = launcher.NewLocalBrowserLauncher
var NewRemoteBrowserLauncher = launcher.NewRemoteBrowserLauncher
var NewBrowserbaseBrowserLauncher = launcher.NewBrowserbaseBrowserLauncher
var NewNoopBrowserLauncher = launcher.NewNoopBrowserLauncher
