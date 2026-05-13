// Code generated from Chrome DevTools Protocol JSON. DO NOT EDIT.
package client

type AccessibilityDomain struct {
	client *ModCDPClient
	On     AccessibilityEvents
}

func (d AccessibilityDomain) Disable(params ...AccessibilityDisableParams) (AccessibilityDisableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return AccessibilityDisableResult{}, err
	}
	return sendCDPCommand[AccessibilityDisableResult](d.client, "Accessibility.disable", p)
}

func (d AccessibilityDomain) Enable(params ...AccessibilityEnableParams) (AccessibilityEnableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return AccessibilityEnableResult{}, err
	}
	return sendCDPCommand[AccessibilityEnableResult](d.client, "Accessibility.enable", p)
}

func (d AccessibilityDomain) GetPartialAXTree(params ...AccessibilityGetPartialAXTreeParams) (AccessibilityGetPartialAXTreeResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return AccessibilityGetPartialAXTreeResult{}, err
	}
	return sendCDPCommand[AccessibilityGetPartialAXTreeResult](d.client, "Accessibility.getPartialAXTree", p)
}

func (d AccessibilityDomain) GetFullAXTree(params ...AccessibilityGetFullAXTreeParams) (AccessibilityGetFullAXTreeResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return AccessibilityGetFullAXTreeResult{}, err
	}
	return sendCDPCommand[AccessibilityGetFullAXTreeResult](d.client, "Accessibility.getFullAXTree", p)
}

func (d AccessibilityDomain) GetRootAXNode(params ...AccessibilityGetRootAXNodeParams) (AccessibilityGetRootAXNodeResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return AccessibilityGetRootAXNodeResult{}, err
	}
	return sendCDPCommand[AccessibilityGetRootAXNodeResult](d.client, "Accessibility.getRootAXNode", p)
}

func (d AccessibilityDomain) GetAXNodeAndAncestors(params ...AccessibilityGetAXNodeAndAncestorsParams) (AccessibilityGetAXNodeAndAncestorsResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return AccessibilityGetAXNodeAndAncestorsResult{}, err
	}
	return sendCDPCommand[AccessibilityGetAXNodeAndAncestorsResult](d.client, "Accessibility.getAXNodeAndAncestors", p)
}

func (d AccessibilityDomain) GetChildAXNodes(params AccessibilityGetChildAXNodesParams) (AccessibilityGetChildAXNodesResult, error) {
	return sendCDPCommand[AccessibilityGetChildAXNodesResult](d.client, "Accessibility.getChildAXNodes", params)
}

func (d AccessibilityDomain) QueryAXTree(params ...AccessibilityQueryAXTreeParams) (AccessibilityQueryAXTreeResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return AccessibilityQueryAXTreeResult{}, err
	}
	return sendCDPCommand[AccessibilityQueryAXTreeResult](d.client, "Accessibility.queryAXTree", p)
}

type AnimationDomain struct {
	client *ModCDPClient
	On     AnimationEvents
}

func (d AnimationDomain) Disable(params ...AnimationDisableParams) (AnimationDisableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return AnimationDisableResult{}, err
	}
	return sendCDPCommand[AnimationDisableResult](d.client, "Animation.disable", p)
}

func (d AnimationDomain) Enable(params ...AnimationEnableParams) (AnimationEnableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return AnimationEnableResult{}, err
	}
	return sendCDPCommand[AnimationEnableResult](d.client, "Animation.enable", p)
}

func (d AnimationDomain) GetCurrentTime(params AnimationGetCurrentTimeParams) (AnimationGetCurrentTimeResult, error) {
	return sendCDPCommand[AnimationGetCurrentTimeResult](d.client, "Animation.getCurrentTime", params)
}

func (d AnimationDomain) GetPlaybackRate(params ...AnimationGetPlaybackRateParams) (AnimationGetPlaybackRateResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return AnimationGetPlaybackRateResult{}, err
	}
	return sendCDPCommand[AnimationGetPlaybackRateResult](d.client, "Animation.getPlaybackRate", p)
}

func (d AnimationDomain) ReleaseAnimations(params AnimationReleaseAnimationsParams) (AnimationReleaseAnimationsResult, error) {
	return sendCDPCommand[AnimationReleaseAnimationsResult](d.client, "Animation.releaseAnimations", params)
}

func (d AnimationDomain) ResolveAnimation(params AnimationResolveAnimationParams) (AnimationResolveAnimationResult, error) {
	return sendCDPCommand[AnimationResolveAnimationResult](d.client, "Animation.resolveAnimation", params)
}

func (d AnimationDomain) SeekAnimations(params AnimationSeekAnimationsParams) (AnimationSeekAnimationsResult, error) {
	return sendCDPCommand[AnimationSeekAnimationsResult](d.client, "Animation.seekAnimations", params)
}

func (d AnimationDomain) SetPaused(params AnimationSetPausedParams) (AnimationSetPausedResult, error) {
	return sendCDPCommand[AnimationSetPausedResult](d.client, "Animation.setPaused", params)
}

func (d AnimationDomain) SetPlaybackRate(params AnimationSetPlaybackRateParams) (AnimationSetPlaybackRateResult, error) {
	return sendCDPCommand[AnimationSetPlaybackRateResult](d.client, "Animation.setPlaybackRate", params)
}

func (d AnimationDomain) SetTiming(params AnimationSetTimingParams) (AnimationSetTimingResult, error) {
	return sendCDPCommand[AnimationSetTimingResult](d.client, "Animation.setTiming", params)
}

type AuditsDomain struct {
	client *ModCDPClient
	On     AuditsEvents
}

func (d AuditsDomain) GetEncodedResponse(params AuditsGetEncodedResponseParams) (AuditsGetEncodedResponseResult, error) {
	return sendCDPCommand[AuditsGetEncodedResponseResult](d.client, "Audits.getEncodedResponse", params)
}

func (d AuditsDomain) Disable(params ...AuditsDisableParams) (AuditsDisableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return AuditsDisableResult{}, err
	}
	return sendCDPCommand[AuditsDisableResult](d.client, "Audits.disable", p)
}

func (d AuditsDomain) Enable(params ...AuditsEnableParams) (AuditsEnableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return AuditsEnableResult{}, err
	}
	return sendCDPCommand[AuditsEnableResult](d.client, "Audits.enable", p)
}

func (d AuditsDomain) CheckFormsIssues(params ...AuditsCheckFormsIssuesParams) (AuditsCheckFormsIssuesResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return AuditsCheckFormsIssuesResult{}, err
	}
	return sendCDPCommand[AuditsCheckFormsIssuesResult](d.client, "Audits.checkFormsIssues", p)
}

type AutofillDomain struct {
	client *ModCDPClient
	On     AutofillEvents
}

func (d AutofillDomain) Trigger(params AutofillTriggerParams) (AutofillTriggerResult, error) {
	return sendCDPCommand[AutofillTriggerResult](d.client, "Autofill.trigger", params)
}

func (d AutofillDomain) SetAddresses(params AutofillSetAddressesParams) (AutofillSetAddressesResult, error) {
	return sendCDPCommand[AutofillSetAddressesResult](d.client, "Autofill.setAddresses", params)
}

func (d AutofillDomain) Disable(params ...AutofillDisableParams) (AutofillDisableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return AutofillDisableResult{}, err
	}
	return sendCDPCommand[AutofillDisableResult](d.client, "Autofill.disable", p)
}

func (d AutofillDomain) Enable(params ...AutofillEnableParams) (AutofillEnableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return AutofillEnableResult{}, err
	}
	return sendCDPCommand[AutofillEnableResult](d.client, "Autofill.enable", p)
}

type BackgroundServiceDomain struct {
	client *ModCDPClient
	On     BackgroundServiceEvents
}

func (d BackgroundServiceDomain) StartObserving(params BackgroundServiceStartObservingParams) (BackgroundServiceStartObservingResult, error) {
	return sendCDPCommand[BackgroundServiceStartObservingResult](d.client, "BackgroundService.startObserving", params)
}

func (d BackgroundServiceDomain) StopObserving(params BackgroundServiceStopObservingParams) (BackgroundServiceStopObservingResult, error) {
	return sendCDPCommand[BackgroundServiceStopObservingResult](d.client, "BackgroundService.stopObserving", params)
}

func (d BackgroundServiceDomain) SetRecording(params BackgroundServiceSetRecordingParams) (BackgroundServiceSetRecordingResult, error) {
	return sendCDPCommand[BackgroundServiceSetRecordingResult](d.client, "BackgroundService.setRecording", params)
}

func (d BackgroundServiceDomain) ClearEvents(params BackgroundServiceClearEventsParams) (BackgroundServiceClearEventsResult, error) {
	return sendCDPCommand[BackgroundServiceClearEventsResult](d.client, "BackgroundService.clearEvents", params)
}

type BluetoothEmulationDomain struct {
	client *ModCDPClient
	On     BluetoothEmulationEvents
}

func (d BluetoothEmulationDomain) Enable(params BluetoothEmulationEnableParams) (BluetoothEmulationEnableResult, error) {
	return sendCDPCommand[BluetoothEmulationEnableResult](d.client, "BluetoothEmulation.enable", params)
}

func (d BluetoothEmulationDomain) SetSimulatedCentralState(params BluetoothEmulationSetSimulatedCentralStateParams) (BluetoothEmulationSetSimulatedCentralStateResult, error) {
	return sendCDPCommand[BluetoothEmulationSetSimulatedCentralStateResult](d.client, "BluetoothEmulation.setSimulatedCentralState", params)
}

func (d BluetoothEmulationDomain) Disable(params ...BluetoothEmulationDisableParams) (BluetoothEmulationDisableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return BluetoothEmulationDisableResult{}, err
	}
	return sendCDPCommand[BluetoothEmulationDisableResult](d.client, "BluetoothEmulation.disable", p)
}

func (d BluetoothEmulationDomain) SimulatePreconnectedPeripheral(params BluetoothEmulationSimulatePreconnectedPeripheralParams) (BluetoothEmulationSimulatePreconnectedPeripheralResult, error) {
	return sendCDPCommand[BluetoothEmulationSimulatePreconnectedPeripheralResult](d.client, "BluetoothEmulation.simulatePreconnectedPeripheral", params)
}

func (d BluetoothEmulationDomain) SimulateAdvertisement(params BluetoothEmulationSimulateAdvertisementParams) (BluetoothEmulationSimulateAdvertisementResult, error) {
	return sendCDPCommand[BluetoothEmulationSimulateAdvertisementResult](d.client, "BluetoothEmulation.simulateAdvertisement", params)
}

func (d BluetoothEmulationDomain) SimulateGATTOperationResponse(params BluetoothEmulationSimulateGATTOperationResponseParams) (BluetoothEmulationSimulateGATTOperationResponseResult, error) {
	return sendCDPCommand[BluetoothEmulationSimulateGATTOperationResponseResult](d.client, "BluetoothEmulation.simulateGATTOperationResponse", params)
}

func (d BluetoothEmulationDomain) SimulateCharacteristicOperationResponse(params BluetoothEmulationSimulateCharacteristicOperationResponseParams) (BluetoothEmulationSimulateCharacteristicOperationResponseResult, error) {
	return sendCDPCommand[BluetoothEmulationSimulateCharacteristicOperationResponseResult](d.client, "BluetoothEmulation.simulateCharacteristicOperationResponse", params)
}

func (d BluetoothEmulationDomain) SimulateDescriptorOperationResponse(params BluetoothEmulationSimulateDescriptorOperationResponseParams) (BluetoothEmulationSimulateDescriptorOperationResponseResult, error) {
	return sendCDPCommand[BluetoothEmulationSimulateDescriptorOperationResponseResult](d.client, "BluetoothEmulation.simulateDescriptorOperationResponse", params)
}

func (d BluetoothEmulationDomain) AddService(params BluetoothEmulationAddServiceParams) (BluetoothEmulationAddServiceResult, error) {
	return sendCDPCommand[BluetoothEmulationAddServiceResult](d.client, "BluetoothEmulation.addService", params)
}

func (d BluetoothEmulationDomain) RemoveService(params BluetoothEmulationRemoveServiceParams) (BluetoothEmulationRemoveServiceResult, error) {
	return sendCDPCommand[BluetoothEmulationRemoveServiceResult](d.client, "BluetoothEmulation.removeService", params)
}

func (d BluetoothEmulationDomain) AddCharacteristic(params BluetoothEmulationAddCharacteristicParams) (BluetoothEmulationAddCharacteristicResult, error) {
	return sendCDPCommand[BluetoothEmulationAddCharacteristicResult](d.client, "BluetoothEmulation.addCharacteristic", params)
}

func (d BluetoothEmulationDomain) RemoveCharacteristic(params BluetoothEmulationRemoveCharacteristicParams) (BluetoothEmulationRemoveCharacteristicResult, error) {
	return sendCDPCommand[BluetoothEmulationRemoveCharacteristicResult](d.client, "BluetoothEmulation.removeCharacteristic", params)
}

func (d BluetoothEmulationDomain) AddDescriptor(params BluetoothEmulationAddDescriptorParams) (BluetoothEmulationAddDescriptorResult, error) {
	return sendCDPCommand[BluetoothEmulationAddDescriptorResult](d.client, "BluetoothEmulation.addDescriptor", params)
}

func (d BluetoothEmulationDomain) RemoveDescriptor(params BluetoothEmulationRemoveDescriptorParams) (BluetoothEmulationRemoveDescriptorResult, error) {
	return sendCDPCommand[BluetoothEmulationRemoveDescriptorResult](d.client, "BluetoothEmulation.removeDescriptor", params)
}

func (d BluetoothEmulationDomain) SimulateGATTDisconnection(params BluetoothEmulationSimulateGATTDisconnectionParams) (BluetoothEmulationSimulateGATTDisconnectionResult, error) {
	return sendCDPCommand[BluetoothEmulationSimulateGATTDisconnectionResult](d.client, "BluetoothEmulation.simulateGATTDisconnection", params)
}

type BrowserDomain struct {
	client *ModCDPClient
	On     BrowserEvents
}

func (d BrowserDomain) SetPermission(params BrowserSetPermissionParams) (BrowserSetPermissionResult, error) {
	return sendCDPCommand[BrowserSetPermissionResult](d.client, "Browser.setPermission", params)
}

func (d BrowserDomain) GrantPermissions(params BrowserGrantPermissionsParams) (BrowserGrantPermissionsResult, error) {
	return sendCDPCommand[BrowserGrantPermissionsResult](d.client, "Browser.grantPermissions", params)
}

func (d BrowserDomain) ResetPermissions(params ...BrowserResetPermissionsParams) (BrowserResetPermissionsResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return BrowserResetPermissionsResult{}, err
	}
	return sendCDPCommand[BrowserResetPermissionsResult](d.client, "Browser.resetPermissions", p)
}

func (d BrowserDomain) SetDownloadBehavior(params BrowserSetDownloadBehaviorParams) (BrowserSetDownloadBehaviorResult, error) {
	return sendCDPCommand[BrowserSetDownloadBehaviorResult](d.client, "Browser.setDownloadBehavior", params)
}

func (d BrowserDomain) CancelDownload(params BrowserCancelDownloadParams) (BrowserCancelDownloadResult, error) {
	return sendCDPCommand[BrowserCancelDownloadResult](d.client, "Browser.cancelDownload", params)
}

func (d BrowserDomain) Close(params ...BrowserCloseParams) (BrowserCloseResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return BrowserCloseResult{}, err
	}
	return sendCDPCommand[BrowserCloseResult](d.client, "Browser.close", p)
}

func (d BrowserDomain) Crash(params ...BrowserCrashParams) (BrowserCrashResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return BrowserCrashResult{}, err
	}
	return sendCDPCommand[BrowserCrashResult](d.client, "Browser.crash", p)
}

func (d BrowserDomain) CrashGPUProcess(params ...BrowserCrashGPUProcessParams) (BrowserCrashGPUProcessResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return BrowserCrashGPUProcessResult{}, err
	}
	return sendCDPCommand[BrowserCrashGPUProcessResult](d.client, "Browser.crashGpuProcess", p)
}

func (d BrowserDomain) GetVersion(params ...BrowserGetVersionParams) (BrowserGetVersionResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return BrowserGetVersionResult{}, err
	}
	return sendCDPCommand[BrowserGetVersionResult](d.client, "Browser.getVersion", p)
}

func (d BrowserDomain) GetBrowserCommandLine(params ...BrowserGetBrowserCommandLineParams) (BrowserGetBrowserCommandLineResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return BrowserGetBrowserCommandLineResult{}, err
	}
	return sendCDPCommand[BrowserGetBrowserCommandLineResult](d.client, "Browser.getBrowserCommandLine", p)
}

func (d BrowserDomain) GetHistograms(params ...BrowserGetHistogramsParams) (BrowserGetHistogramsResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return BrowserGetHistogramsResult{}, err
	}
	return sendCDPCommand[BrowserGetHistogramsResult](d.client, "Browser.getHistograms", p)
}

func (d BrowserDomain) GetHistogram(params BrowserGetHistogramParams) (BrowserGetHistogramResult, error) {
	return sendCDPCommand[BrowserGetHistogramResult](d.client, "Browser.getHistogram", params)
}

func (d BrowserDomain) GetWindowBounds(params BrowserGetWindowBoundsParams) (BrowserGetWindowBoundsResult, error) {
	return sendCDPCommand[BrowserGetWindowBoundsResult](d.client, "Browser.getWindowBounds", params)
}

func (d BrowserDomain) GetWindowForTarget(params ...BrowserGetWindowForTargetParams) (BrowserGetWindowForTargetResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return BrowserGetWindowForTargetResult{}, err
	}
	return sendCDPCommand[BrowserGetWindowForTargetResult](d.client, "Browser.getWindowForTarget", p)
}

func (d BrowserDomain) SetWindowBounds(params BrowserSetWindowBoundsParams) (BrowserSetWindowBoundsResult, error) {
	return sendCDPCommand[BrowserSetWindowBoundsResult](d.client, "Browser.setWindowBounds", params)
}

func (d BrowserDomain) SetContentsSize(params BrowserSetContentsSizeParams) (BrowserSetContentsSizeResult, error) {
	return sendCDPCommand[BrowserSetContentsSizeResult](d.client, "Browser.setContentsSize", params)
}

func (d BrowserDomain) SetDockTile(params ...BrowserSetDockTileParams) (BrowserSetDockTileResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return BrowserSetDockTileResult{}, err
	}
	return sendCDPCommand[BrowserSetDockTileResult](d.client, "Browser.setDockTile", p)
}

func (d BrowserDomain) ExecuteBrowserCommand(params BrowserExecuteBrowserCommandParams) (BrowserExecuteBrowserCommandResult, error) {
	return sendCDPCommand[BrowserExecuteBrowserCommandResult](d.client, "Browser.executeBrowserCommand", params)
}

func (d BrowserDomain) AddPrivacySandboxEnrollmentOverride(params BrowserAddPrivacySandboxEnrollmentOverrideParams) (BrowserAddPrivacySandboxEnrollmentOverrideResult, error) {
	return sendCDPCommand[BrowserAddPrivacySandboxEnrollmentOverrideResult](d.client, "Browser.addPrivacySandboxEnrollmentOverride", params)
}

func (d BrowserDomain) AddPrivacySandboxCoordinatorKeyConfig(params BrowserAddPrivacySandboxCoordinatorKeyConfigParams) (BrowserAddPrivacySandboxCoordinatorKeyConfigResult, error) {
	return sendCDPCommand[BrowserAddPrivacySandboxCoordinatorKeyConfigResult](d.client, "Browser.addPrivacySandboxCoordinatorKeyConfig", params)
}

type CSSDomain struct {
	client *ModCDPClient
	On     CSSEvents
}

func (d CSSDomain) AddRule(params CSSAddRuleParams) (CSSAddRuleResult, error) {
	return sendCDPCommand[CSSAddRuleResult](d.client, "CSS.addRule", params)
}

func (d CSSDomain) CollectClassNames(params CSSCollectClassNamesParams) (CSSCollectClassNamesResult, error) {
	return sendCDPCommand[CSSCollectClassNamesResult](d.client, "CSS.collectClassNames", params)
}

func (d CSSDomain) CreateStyleSheet(params CSSCreateStyleSheetParams) (CSSCreateStyleSheetResult, error) {
	return sendCDPCommand[CSSCreateStyleSheetResult](d.client, "CSS.createStyleSheet", params)
}

func (d CSSDomain) Disable(params ...CSSDisableParams) (CSSDisableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return CSSDisableResult{}, err
	}
	return sendCDPCommand[CSSDisableResult](d.client, "CSS.disable", p)
}

func (d CSSDomain) Enable(params ...CSSEnableParams) (CSSEnableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return CSSEnableResult{}, err
	}
	return sendCDPCommand[CSSEnableResult](d.client, "CSS.enable", p)
}

func (d CSSDomain) ForcePseudoState(params CSSForcePseudoStateParams) (CSSForcePseudoStateResult, error) {
	return sendCDPCommand[CSSForcePseudoStateResult](d.client, "CSS.forcePseudoState", params)
}

func (d CSSDomain) ForceStartingStyle(params CSSForceStartingStyleParams) (CSSForceStartingStyleResult, error) {
	return sendCDPCommand[CSSForceStartingStyleResult](d.client, "CSS.forceStartingStyle", params)
}

func (d CSSDomain) GetBackgroundColors(params CSSGetBackgroundColorsParams) (CSSGetBackgroundColorsResult, error) {
	return sendCDPCommand[CSSGetBackgroundColorsResult](d.client, "CSS.getBackgroundColors", params)
}

func (d CSSDomain) GetComputedStyleForNode(params CSSGetComputedStyleForNodeParams) (CSSGetComputedStyleForNodeResult, error) {
	return sendCDPCommand[CSSGetComputedStyleForNodeResult](d.client, "CSS.getComputedStyleForNode", params)
}

func (d CSSDomain) ResolveValues(params CSSResolveValuesParams) (CSSResolveValuesResult, error) {
	return sendCDPCommand[CSSResolveValuesResult](d.client, "CSS.resolveValues", params)
}

func (d CSSDomain) GetLonghandProperties(params CSSGetLonghandPropertiesParams) (CSSGetLonghandPropertiesResult, error) {
	return sendCDPCommand[CSSGetLonghandPropertiesResult](d.client, "CSS.getLonghandProperties", params)
}

func (d CSSDomain) GetInlineStylesForNode(params CSSGetInlineStylesForNodeParams) (CSSGetInlineStylesForNodeResult, error) {
	return sendCDPCommand[CSSGetInlineStylesForNodeResult](d.client, "CSS.getInlineStylesForNode", params)
}

func (d CSSDomain) GetAnimatedStylesForNode(params CSSGetAnimatedStylesForNodeParams) (CSSGetAnimatedStylesForNodeResult, error) {
	return sendCDPCommand[CSSGetAnimatedStylesForNodeResult](d.client, "CSS.getAnimatedStylesForNode", params)
}

func (d CSSDomain) GetMatchedStylesForNode(params CSSGetMatchedStylesForNodeParams) (CSSGetMatchedStylesForNodeResult, error) {
	return sendCDPCommand[CSSGetMatchedStylesForNodeResult](d.client, "CSS.getMatchedStylesForNode", params)
}

func (d CSSDomain) GetEnvironmentVariables(params ...CSSGetEnvironmentVariablesParams) (CSSGetEnvironmentVariablesResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return CSSGetEnvironmentVariablesResult{}, err
	}
	return sendCDPCommand[CSSGetEnvironmentVariablesResult](d.client, "CSS.getEnvironmentVariables", p)
}

func (d CSSDomain) GetMediaQueries(params ...CSSGetMediaQueriesParams) (CSSGetMediaQueriesResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return CSSGetMediaQueriesResult{}, err
	}
	return sendCDPCommand[CSSGetMediaQueriesResult](d.client, "CSS.getMediaQueries", p)
}

func (d CSSDomain) GetPlatformFontsForNode(params CSSGetPlatformFontsForNodeParams) (CSSGetPlatformFontsForNodeResult, error) {
	return sendCDPCommand[CSSGetPlatformFontsForNodeResult](d.client, "CSS.getPlatformFontsForNode", params)
}

func (d CSSDomain) GetStyleSheetText(params CSSGetStyleSheetTextParams) (CSSGetStyleSheetTextResult, error) {
	return sendCDPCommand[CSSGetStyleSheetTextResult](d.client, "CSS.getStyleSheetText", params)
}

func (d CSSDomain) GetLayersForNode(params CSSGetLayersForNodeParams) (CSSGetLayersForNodeResult, error) {
	return sendCDPCommand[CSSGetLayersForNodeResult](d.client, "CSS.getLayersForNode", params)
}

func (d CSSDomain) GetLocationForSelector(params CSSGetLocationForSelectorParams) (CSSGetLocationForSelectorResult, error) {
	return sendCDPCommand[CSSGetLocationForSelectorResult](d.client, "CSS.getLocationForSelector", params)
}

func (d CSSDomain) TrackComputedStyleUpdatesForNode(params ...CSSTrackComputedStyleUpdatesForNodeParams) (CSSTrackComputedStyleUpdatesForNodeResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return CSSTrackComputedStyleUpdatesForNodeResult{}, err
	}
	return sendCDPCommand[CSSTrackComputedStyleUpdatesForNodeResult](d.client, "CSS.trackComputedStyleUpdatesForNode", p)
}

func (d CSSDomain) TrackComputedStyleUpdates(params CSSTrackComputedStyleUpdatesParams) (CSSTrackComputedStyleUpdatesResult, error) {
	return sendCDPCommand[CSSTrackComputedStyleUpdatesResult](d.client, "CSS.trackComputedStyleUpdates", params)
}

func (d CSSDomain) TakeComputedStyleUpdates(params ...CSSTakeComputedStyleUpdatesParams) (CSSTakeComputedStyleUpdatesResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return CSSTakeComputedStyleUpdatesResult{}, err
	}
	return sendCDPCommand[CSSTakeComputedStyleUpdatesResult](d.client, "CSS.takeComputedStyleUpdates", p)
}

func (d CSSDomain) SetEffectivePropertyValueForNode(params CSSSetEffectivePropertyValueForNodeParams) (CSSSetEffectivePropertyValueForNodeResult, error) {
	return sendCDPCommand[CSSSetEffectivePropertyValueForNodeResult](d.client, "CSS.setEffectivePropertyValueForNode", params)
}

func (d CSSDomain) SetPropertyRulePropertyName(params CSSSetPropertyRulePropertyNameParams) (CSSSetPropertyRulePropertyNameResult, error) {
	return sendCDPCommand[CSSSetPropertyRulePropertyNameResult](d.client, "CSS.setPropertyRulePropertyName", params)
}

func (d CSSDomain) SetKeyframeKey(params CSSSetKeyframeKeyParams) (CSSSetKeyframeKeyResult, error) {
	return sendCDPCommand[CSSSetKeyframeKeyResult](d.client, "CSS.setKeyframeKey", params)
}

func (d CSSDomain) SetMediaText(params CSSSetMediaTextParams) (CSSSetMediaTextResult, error) {
	return sendCDPCommand[CSSSetMediaTextResult](d.client, "CSS.setMediaText", params)
}

func (d CSSDomain) SetContainerQueryText(params CSSSetContainerQueryTextParams) (CSSSetContainerQueryTextResult, error) {
	return sendCDPCommand[CSSSetContainerQueryTextResult](d.client, "CSS.setContainerQueryText", params)
}

func (d CSSDomain) SetSupportsText(params CSSSetSupportsTextParams) (CSSSetSupportsTextResult, error) {
	return sendCDPCommand[CSSSetSupportsTextResult](d.client, "CSS.setSupportsText", params)
}

func (d CSSDomain) SetNavigationText(params CSSSetNavigationTextParams) (CSSSetNavigationTextResult, error) {
	return sendCDPCommand[CSSSetNavigationTextResult](d.client, "CSS.setNavigationText", params)
}

func (d CSSDomain) SetScopeText(params CSSSetScopeTextParams) (CSSSetScopeTextResult, error) {
	return sendCDPCommand[CSSSetScopeTextResult](d.client, "CSS.setScopeText", params)
}

func (d CSSDomain) SetRuleSelector(params CSSSetRuleSelectorParams) (CSSSetRuleSelectorResult, error) {
	return sendCDPCommand[CSSSetRuleSelectorResult](d.client, "CSS.setRuleSelector", params)
}

func (d CSSDomain) SetStyleSheetText(params CSSSetStyleSheetTextParams) (CSSSetStyleSheetTextResult, error) {
	return sendCDPCommand[CSSSetStyleSheetTextResult](d.client, "CSS.setStyleSheetText", params)
}

func (d CSSDomain) SetStyleTexts(params CSSSetStyleTextsParams) (CSSSetStyleTextsResult, error) {
	return sendCDPCommand[CSSSetStyleTextsResult](d.client, "CSS.setStyleTexts", params)
}

func (d CSSDomain) StartRuleUsageTracking(params ...CSSStartRuleUsageTrackingParams) (CSSStartRuleUsageTrackingResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return CSSStartRuleUsageTrackingResult{}, err
	}
	return sendCDPCommand[CSSStartRuleUsageTrackingResult](d.client, "CSS.startRuleUsageTracking", p)
}

func (d CSSDomain) StopRuleUsageTracking(params ...CSSStopRuleUsageTrackingParams) (CSSStopRuleUsageTrackingResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return CSSStopRuleUsageTrackingResult{}, err
	}
	return sendCDPCommand[CSSStopRuleUsageTrackingResult](d.client, "CSS.stopRuleUsageTracking", p)
}

func (d CSSDomain) TakeCoverageDelta(params ...CSSTakeCoverageDeltaParams) (CSSTakeCoverageDeltaResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return CSSTakeCoverageDeltaResult{}, err
	}
	return sendCDPCommand[CSSTakeCoverageDeltaResult](d.client, "CSS.takeCoverageDelta", p)
}

func (d CSSDomain) SetLocalFontsEnabled(params CSSSetLocalFontsEnabledParams) (CSSSetLocalFontsEnabledResult, error) {
	return sendCDPCommand[CSSSetLocalFontsEnabledResult](d.client, "CSS.setLocalFontsEnabled", params)
}

type CacheStorageDomain struct{ client *ModCDPClient }

func (d CacheStorageDomain) DeleteCache(params CacheStorageDeleteCacheParams) (CacheStorageDeleteCacheResult, error) {
	return sendCDPCommand[CacheStorageDeleteCacheResult](d.client, "CacheStorage.deleteCache", params)
}

func (d CacheStorageDomain) DeleteEntry(params CacheStorageDeleteEntryParams) (CacheStorageDeleteEntryResult, error) {
	return sendCDPCommand[CacheStorageDeleteEntryResult](d.client, "CacheStorage.deleteEntry", params)
}

func (d CacheStorageDomain) RequestCacheNames(params ...CacheStorageRequestCacheNamesParams) (CacheStorageRequestCacheNamesResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return CacheStorageRequestCacheNamesResult{}, err
	}
	return sendCDPCommand[CacheStorageRequestCacheNamesResult](d.client, "CacheStorage.requestCacheNames", p)
}

func (d CacheStorageDomain) RequestCachedResponse(params CacheStorageRequestCachedResponseParams) (CacheStorageRequestCachedResponseResult, error) {
	return sendCDPCommand[CacheStorageRequestCachedResponseResult](d.client, "CacheStorage.requestCachedResponse", params)
}

func (d CacheStorageDomain) RequestEntries(params CacheStorageRequestEntriesParams) (CacheStorageRequestEntriesResult, error) {
	return sendCDPCommand[CacheStorageRequestEntriesResult](d.client, "CacheStorage.requestEntries", params)
}

type CastDomain struct {
	client *ModCDPClient
	On     CastEvents
}

func (d CastDomain) Enable(params ...CastEnableParams) (CastEnableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return CastEnableResult{}, err
	}
	return sendCDPCommand[CastEnableResult](d.client, "Cast.enable", p)
}

func (d CastDomain) Disable(params ...CastDisableParams) (CastDisableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return CastDisableResult{}, err
	}
	return sendCDPCommand[CastDisableResult](d.client, "Cast.disable", p)
}

func (d CastDomain) SetSinkToUse(params CastSetSinkToUseParams) (CastSetSinkToUseResult, error) {
	return sendCDPCommand[CastSetSinkToUseResult](d.client, "Cast.setSinkToUse", params)
}

func (d CastDomain) StartDesktopMirroring(params CastStartDesktopMirroringParams) (CastStartDesktopMirroringResult, error) {
	return sendCDPCommand[CastStartDesktopMirroringResult](d.client, "Cast.startDesktopMirroring", params)
}

func (d CastDomain) StartTabMirroring(params CastStartTabMirroringParams) (CastStartTabMirroringResult, error) {
	return sendCDPCommand[CastStartTabMirroringResult](d.client, "Cast.startTabMirroring", params)
}

func (d CastDomain) StopCasting(params CastStopCastingParams) (CastStopCastingResult, error) {
	return sendCDPCommand[CastStopCastingResult](d.client, "Cast.stopCasting", params)
}

type ConsoleDomain struct {
	client *ModCDPClient
	On     ConsoleEvents
}

func (d ConsoleDomain) ClearMessages(params ...ConsoleClearMessagesParams) (ConsoleClearMessagesResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return ConsoleClearMessagesResult{}, err
	}
	return sendCDPCommand[ConsoleClearMessagesResult](d.client, "Console.clearMessages", p)
}

func (d ConsoleDomain) Disable(params ...ConsoleDisableParams) (ConsoleDisableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return ConsoleDisableResult{}, err
	}
	return sendCDPCommand[ConsoleDisableResult](d.client, "Console.disable", p)
}

func (d ConsoleDomain) Enable(params ...ConsoleEnableParams) (ConsoleEnableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return ConsoleEnableResult{}, err
	}
	return sendCDPCommand[ConsoleEnableResult](d.client, "Console.enable", p)
}

type DOMDomain struct {
	client *ModCDPClient
	On     DOMEvents
}

func (d DOMDomain) CollectClassNamesFromSubtree(params DOMCollectClassNamesFromSubtreeParams) (DOMCollectClassNamesFromSubtreeResult, error) {
	return sendCDPCommand[DOMCollectClassNamesFromSubtreeResult](d.client, "DOM.collectClassNamesFromSubtree", params)
}

func (d DOMDomain) CopyTo(params DOMCopyToParams) (DOMCopyToResult, error) {
	return sendCDPCommand[DOMCopyToResult](d.client, "DOM.copyTo", params)
}

func (d DOMDomain) DescribeNode(params ...DOMDescribeNodeParams) (DOMDescribeNodeResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return DOMDescribeNodeResult{}, err
	}
	return sendCDPCommand[DOMDescribeNodeResult](d.client, "DOM.describeNode", p)
}

func (d DOMDomain) ScrollIntoViewIfNeeded(params ...DOMScrollIntoViewIfNeededParams) (DOMScrollIntoViewIfNeededResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return DOMScrollIntoViewIfNeededResult{}, err
	}
	return sendCDPCommand[DOMScrollIntoViewIfNeededResult](d.client, "DOM.scrollIntoViewIfNeeded", p)
}

func (d DOMDomain) Disable(params ...DOMDisableParams) (DOMDisableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return DOMDisableResult{}, err
	}
	return sendCDPCommand[DOMDisableResult](d.client, "DOM.disable", p)
}

func (d DOMDomain) DiscardSearchResults(params DOMDiscardSearchResultsParams) (DOMDiscardSearchResultsResult, error) {
	return sendCDPCommand[DOMDiscardSearchResultsResult](d.client, "DOM.discardSearchResults", params)
}

func (d DOMDomain) Enable(params ...DOMEnableParams) (DOMEnableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return DOMEnableResult{}, err
	}
	return sendCDPCommand[DOMEnableResult](d.client, "DOM.enable", p)
}

func (d DOMDomain) Focus(params ...DOMFocusParams) (DOMFocusResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return DOMFocusResult{}, err
	}
	return sendCDPCommand[DOMFocusResult](d.client, "DOM.focus", p)
}

func (d DOMDomain) GetAttributes(params DOMGetAttributesParams) (DOMGetAttributesResult, error) {
	return sendCDPCommand[DOMGetAttributesResult](d.client, "DOM.getAttributes", params)
}

func (d DOMDomain) GetBoxModel(params ...DOMGetBoxModelParams) (DOMGetBoxModelResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return DOMGetBoxModelResult{}, err
	}
	return sendCDPCommand[DOMGetBoxModelResult](d.client, "DOM.getBoxModel", p)
}

func (d DOMDomain) GetContentQuads(params ...DOMGetContentQuadsParams) (DOMGetContentQuadsResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return DOMGetContentQuadsResult{}, err
	}
	return sendCDPCommand[DOMGetContentQuadsResult](d.client, "DOM.getContentQuads", p)
}

func (d DOMDomain) GetDocument(params ...DOMGetDocumentParams) (DOMGetDocumentResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return DOMGetDocumentResult{}, err
	}
	return sendCDPCommand[DOMGetDocumentResult](d.client, "DOM.getDocument", p)
}

func (d DOMDomain) GetFlattenedDocument(params ...DOMGetFlattenedDocumentParams) (DOMGetFlattenedDocumentResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return DOMGetFlattenedDocumentResult{}, err
	}
	return sendCDPCommand[DOMGetFlattenedDocumentResult](d.client, "DOM.getFlattenedDocument", p)
}

func (d DOMDomain) GetNodesForSubtreeByStyle(params DOMGetNodesForSubtreeByStyleParams) (DOMGetNodesForSubtreeByStyleResult, error) {
	return sendCDPCommand[DOMGetNodesForSubtreeByStyleResult](d.client, "DOM.getNodesForSubtreeByStyle", params)
}

func (d DOMDomain) GetNodeForLocation(params DOMGetNodeForLocationParams) (DOMGetNodeForLocationResult, error) {
	return sendCDPCommand[DOMGetNodeForLocationResult](d.client, "DOM.getNodeForLocation", params)
}

func (d DOMDomain) GetOuterHTML(params ...DOMGetOuterHTMLParams) (DOMGetOuterHTMLResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return DOMGetOuterHTMLResult{}, err
	}
	return sendCDPCommand[DOMGetOuterHTMLResult](d.client, "DOM.getOuterHTML", p)
}

func (d DOMDomain) GetRelayoutBoundary(params DOMGetRelayoutBoundaryParams) (DOMGetRelayoutBoundaryResult, error) {
	return sendCDPCommand[DOMGetRelayoutBoundaryResult](d.client, "DOM.getRelayoutBoundary", params)
}

func (d DOMDomain) GetSearchResults(params DOMGetSearchResultsParams) (DOMGetSearchResultsResult, error) {
	return sendCDPCommand[DOMGetSearchResultsResult](d.client, "DOM.getSearchResults", params)
}

func (d DOMDomain) HideHighlight(params ...DOMHideHighlightParams) (DOMHideHighlightResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return DOMHideHighlightResult{}, err
	}
	return sendCDPCommand[DOMHideHighlightResult](d.client, "DOM.hideHighlight", p)
}

func (d DOMDomain) HighlightNode(params ...DOMHighlightNodeParams) (DOMHighlightNodeResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return DOMHighlightNodeResult{}, err
	}
	return sendCDPCommand[DOMHighlightNodeResult](d.client, "DOM.highlightNode", p)
}

func (d DOMDomain) HighlightRect(params ...DOMHighlightRectParams) (DOMHighlightRectResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return DOMHighlightRectResult{}, err
	}
	return sendCDPCommand[DOMHighlightRectResult](d.client, "DOM.highlightRect", p)
}

func (d DOMDomain) MarkUndoableState(params ...DOMMarkUndoableStateParams) (DOMMarkUndoableStateResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return DOMMarkUndoableStateResult{}, err
	}
	return sendCDPCommand[DOMMarkUndoableStateResult](d.client, "DOM.markUndoableState", p)
}

func (d DOMDomain) MoveTo(params DOMMoveToParams) (DOMMoveToResult, error) {
	return sendCDPCommand[DOMMoveToResult](d.client, "DOM.moveTo", params)
}

func (d DOMDomain) PerformSearch(params DOMPerformSearchParams) (DOMPerformSearchResult, error) {
	return sendCDPCommand[DOMPerformSearchResult](d.client, "DOM.performSearch", params)
}

func (d DOMDomain) PushNodeByPathToFrontend(params DOMPushNodeByPathToFrontendParams) (DOMPushNodeByPathToFrontendResult, error) {
	return sendCDPCommand[DOMPushNodeByPathToFrontendResult](d.client, "DOM.pushNodeByPathToFrontend", params)
}

func (d DOMDomain) PushNodesByBackendIdsToFrontend(params DOMPushNodesByBackendIdsToFrontendParams) (DOMPushNodesByBackendIdsToFrontendResult, error) {
	return sendCDPCommand[DOMPushNodesByBackendIdsToFrontendResult](d.client, "DOM.pushNodesByBackendIdsToFrontend", params)
}

func (d DOMDomain) QuerySelector(params DOMQuerySelectorParams) (DOMQuerySelectorResult, error) {
	return sendCDPCommand[DOMQuerySelectorResult](d.client, "DOM.querySelector", params)
}

func (d DOMDomain) QuerySelectorAll(params DOMQuerySelectorAllParams) (DOMQuerySelectorAllResult, error) {
	return sendCDPCommand[DOMQuerySelectorAllResult](d.client, "DOM.querySelectorAll", params)
}

func (d DOMDomain) GetTopLayerElements(params ...DOMGetTopLayerElementsParams) (DOMGetTopLayerElementsResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return DOMGetTopLayerElementsResult{}, err
	}
	return sendCDPCommand[DOMGetTopLayerElementsResult](d.client, "DOM.getTopLayerElements", p)
}

func (d DOMDomain) GetElementByRelation(params DOMGetElementByRelationParams) (DOMGetElementByRelationResult, error) {
	return sendCDPCommand[DOMGetElementByRelationResult](d.client, "DOM.getElementByRelation", params)
}

func (d DOMDomain) Redo(params ...DOMRedoParams) (DOMRedoResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return DOMRedoResult{}, err
	}
	return sendCDPCommand[DOMRedoResult](d.client, "DOM.redo", p)
}

func (d DOMDomain) RemoveAttribute(params DOMRemoveAttributeParams) (DOMRemoveAttributeResult, error) {
	return sendCDPCommand[DOMRemoveAttributeResult](d.client, "DOM.removeAttribute", params)
}

func (d DOMDomain) RemoveNode(params DOMRemoveNodeParams) (DOMRemoveNodeResult, error) {
	return sendCDPCommand[DOMRemoveNodeResult](d.client, "DOM.removeNode", params)
}

func (d DOMDomain) RequestChildNodes(params DOMRequestChildNodesParams) (DOMRequestChildNodesResult, error) {
	return sendCDPCommand[DOMRequestChildNodesResult](d.client, "DOM.requestChildNodes", params)
}

func (d DOMDomain) RequestNode(params DOMRequestNodeParams) (DOMRequestNodeResult, error) {
	return sendCDPCommand[DOMRequestNodeResult](d.client, "DOM.requestNode", params)
}

func (d DOMDomain) ResolveNode(params ...DOMResolveNodeParams) (DOMResolveNodeResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return DOMResolveNodeResult{}, err
	}
	return sendCDPCommand[DOMResolveNodeResult](d.client, "DOM.resolveNode", p)
}

func (d DOMDomain) SetAttributeValue(params DOMSetAttributeValueParams) (DOMSetAttributeValueResult, error) {
	return sendCDPCommand[DOMSetAttributeValueResult](d.client, "DOM.setAttributeValue", params)
}

func (d DOMDomain) SetAttributesAsText(params DOMSetAttributesAsTextParams) (DOMSetAttributesAsTextResult, error) {
	return sendCDPCommand[DOMSetAttributesAsTextResult](d.client, "DOM.setAttributesAsText", params)
}

func (d DOMDomain) SetFileInputFiles(params DOMSetFileInputFilesParams) (DOMSetFileInputFilesResult, error) {
	return sendCDPCommand[DOMSetFileInputFilesResult](d.client, "DOM.setFileInputFiles", params)
}

func (d DOMDomain) SetNodeStackTracesEnabled(params DOMSetNodeStackTracesEnabledParams) (DOMSetNodeStackTracesEnabledResult, error) {
	return sendCDPCommand[DOMSetNodeStackTracesEnabledResult](d.client, "DOM.setNodeStackTracesEnabled", params)
}

func (d DOMDomain) GetNodeStackTraces(params DOMGetNodeStackTracesParams) (DOMGetNodeStackTracesResult, error) {
	return sendCDPCommand[DOMGetNodeStackTracesResult](d.client, "DOM.getNodeStackTraces", params)
}

func (d DOMDomain) GetFileInfo(params DOMGetFileInfoParams) (DOMGetFileInfoResult, error) {
	return sendCDPCommand[DOMGetFileInfoResult](d.client, "DOM.getFileInfo", params)
}

func (d DOMDomain) GetDetachedDOMNodes(params ...DOMGetDetachedDOMNodesParams) (DOMGetDetachedDOMNodesResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return DOMGetDetachedDOMNodesResult{}, err
	}
	return sendCDPCommand[DOMGetDetachedDOMNodesResult](d.client, "DOM.getDetachedDomNodes", p)
}

func (d DOMDomain) SetInspectedNode(params DOMSetInspectedNodeParams) (DOMSetInspectedNodeResult, error) {
	return sendCDPCommand[DOMSetInspectedNodeResult](d.client, "DOM.setInspectedNode", params)
}

func (d DOMDomain) SetNodeName(params DOMSetNodeNameParams) (DOMSetNodeNameResult, error) {
	return sendCDPCommand[DOMSetNodeNameResult](d.client, "DOM.setNodeName", params)
}

func (d DOMDomain) SetNodeValue(params DOMSetNodeValueParams) (DOMSetNodeValueResult, error) {
	return sendCDPCommand[DOMSetNodeValueResult](d.client, "DOM.setNodeValue", params)
}

func (d DOMDomain) SetOuterHTML(params DOMSetOuterHTMLParams) (DOMSetOuterHTMLResult, error) {
	return sendCDPCommand[DOMSetOuterHTMLResult](d.client, "DOM.setOuterHTML", params)
}

func (d DOMDomain) Undo(params ...DOMUndoParams) (DOMUndoResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return DOMUndoResult{}, err
	}
	return sendCDPCommand[DOMUndoResult](d.client, "DOM.undo", p)
}

func (d DOMDomain) GetFrameOwner(params DOMGetFrameOwnerParams) (DOMGetFrameOwnerResult, error) {
	return sendCDPCommand[DOMGetFrameOwnerResult](d.client, "DOM.getFrameOwner", params)
}

func (d DOMDomain) GetContainerForNode(params DOMGetContainerForNodeParams) (DOMGetContainerForNodeResult, error) {
	return sendCDPCommand[DOMGetContainerForNodeResult](d.client, "DOM.getContainerForNode", params)
}

func (d DOMDomain) GetQueryingDescendantsForContainer(params DOMGetQueryingDescendantsForContainerParams) (DOMGetQueryingDescendantsForContainerResult, error) {
	return sendCDPCommand[DOMGetQueryingDescendantsForContainerResult](d.client, "DOM.getQueryingDescendantsForContainer", params)
}

func (d DOMDomain) GetAnchorElement(params DOMGetAnchorElementParams) (DOMGetAnchorElementResult, error) {
	return sendCDPCommand[DOMGetAnchorElementResult](d.client, "DOM.getAnchorElement", params)
}

func (d DOMDomain) ForceShowPopover(params DOMForceShowPopoverParams) (DOMForceShowPopoverResult, error) {
	return sendCDPCommand[DOMForceShowPopoverResult](d.client, "DOM.forceShowPopover", params)
}

type DOMDebuggerDomain struct{ client *ModCDPClient }

func (d DOMDebuggerDomain) GetEventListeners(params DOMDebuggerGetEventListenersParams) (DOMDebuggerGetEventListenersResult, error) {
	return sendCDPCommand[DOMDebuggerGetEventListenersResult](d.client, "DOMDebugger.getEventListeners", params)
}

func (d DOMDebuggerDomain) RemoveDOMBreakpoint(params DOMDebuggerRemoveDOMBreakpointParams) (DOMDebuggerRemoveDOMBreakpointResult, error) {
	return sendCDPCommand[DOMDebuggerRemoveDOMBreakpointResult](d.client, "DOMDebugger.removeDOMBreakpoint", params)
}

func (d DOMDebuggerDomain) RemoveEventListenerBreakpoint(params DOMDebuggerRemoveEventListenerBreakpointParams) (DOMDebuggerRemoveEventListenerBreakpointResult, error) {
	return sendCDPCommand[DOMDebuggerRemoveEventListenerBreakpointResult](d.client, "DOMDebugger.removeEventListenerBreakpoint", params)
}

func (d DOMDebuggerDomain) RemoveInstrumentationBreakpoint(params DOMDebuggerRemoveInstrumentationBreakpointParams) (DOMDebuggerRemoveInstrumentationBreakpointResult, error) {
	return sendCDPCommand[DOMDebuggerRemoveInstrumentationBreakpointResult](d.client, "DOMDebugger.removeInstrumentationBreakpoint", params)
}

func (d DOMDebuggerDomain) RemoveXHRBreakpoint(params DOMDebuggerRemoveXHRBreakpointParams) (DOMDebuggerRemoveXHRBreakpointResult, error) {
	return sendCDPCommand[DOMDebuggerRemoveXHRBreakpointResult](d.client, "DOMDebugger.removeXHRBreakpoint", params)
}

func (d DOMDebuggerDomain) SetBreakOnCSPViolation(params DOMDebuggerSetBreakOnCSPViolationParams) (DOMDebuggerSetBreakOnCSPViolationResult, error) {
	return sendCDPCommand[DOMDebuggerSetBreakOnCSPViolationResult](d.client, "DOMDebugger.setBreakOnCSPViolation", params)
}

func (d DOMDebuggerDomain) SetDOMBreakpoint(params DOMDebuggerSetDOMBreakpointParams) (DOMDebuggerSetDOMBreakpointResult, error) {
	return sendCDPCommand[DOMDebuggerSetDOMBreakpointResult](d.client, "DOMDebugger.setDOMBreakpoint", params)
}

func (d DOMDebuggerDomain) SetEventListenerBreakpoint(params DOMDebuggerSetEventListenerBreakpointParams) (DOMDebuggerSetEventListenerBreakpointResult, error) {
	return sendCDPCommand[DOMDebuggerSetEventListenerBreakpointResult](d.client, "DOMDebugger.setEventListenerBreakpoint", params)
}

func (d DOMDebuggerDomain) SetInstrumentationBreakpoint(params DOMDebuggerSetInstrumentationBreakpointParams) (DOMDebuggerSetInstrumentationBreakpointResult, error) {
	return sendCDPCommand[DOMDebuggerSetInstrumentationBreakpointResult](d.client, "DOMDebugger.setInstrumentationBreakpoint", params)
}

func (d DOMDebuggerDomain) SetXHRBreakpoint(params DOMDebuggerSetXHRBreakpointParams) (DOMDebuggerSetXHRBreakpointResult, error) {
	return sendCDPCommand[DOMDebuggerSetXHRBreakpointResult](d.client, "DOMDebugger.setXHRBreakpoint", params)
}

type DOMSnapshotDomain struct{ client *ModCDPClient }

func (d DOMSnapshotDomain) Disable(params ...DOMSnapshotDisableParams) (DOMSnapshotDisableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return DOMSnapshotDisableResult{}, err
	}
	return sendCDPCommand[DOMSnapshotDisableResult](d.client, "DOMSnapshot.disable", p)
}

func (d DOMSnapshotDomain) Enable(params ...DOMSnapshotEnableParams) (DOMSnapshotEnableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return DOMSnapshotEnableResult{}, err
	}
	return sendCDPCommand[DOMSnapshotEnableResult](d.client, "DOMSnapshot.enable", p)
}

func (d DOMSnapshotDomain) GetSnapshot(params DOMSnapshotGetSnapshotParams) (DOMSnapshotGetSnapshotResult, error) {
	return sendCDPCommand[DOMSnapshotGetSnapshotResult](d.client, "DOMSnapshot.getSnapshot", params)
}

func (d DOMSnapshotDomain) CaptureSnapshot(params DOMSnapshotCaptureSnapshotParams) (DOMSnapshotCaptureSnapshotResult, error) {
	return sendCDPCommand[DOMSnapshotCaptureSnapshotResult](d.client, "DOMSnapshot.captureSnapshot", params)
}

type DOMStorageDomain struct {
	client *ModCDPClient
	On     DOMStorageEvents
}

func (d DOMStorageDomain) Clear(params DOMStorageClearParams) (DOMStorageClearResult, error) {
	return sendCDPCommand[DOMStorageClearResult](d.client, "DOMStorage.clear", params)
}

func (d DOMStorageDomain) Disable(params ...DOMStorageDisableParams) (DOMStorageDisableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return DOMStorageDisableResult{}, err
	}
	return sendCDPCommand[DOMStorageDisableResult](d.client, "DOMStorage.disable", p)
}

func (d DOMStorageDomain) Enable(params ...DOMStorageEnableParams) (DOMStorageEnableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return DOMStorageEnableResult{}, err
	}
	return sendCDPCommand[DOMStorageEnableResult](d.client, "DOMStorage.enable", p)
}

func (d DOMStorageDomain) GetDOMStorageItems(params DOMStorageGetDOMStorageItemsParams) (DOMStorageGetDOMStorageItemsResult, error) {
	return sendCDPCommand[DOMStorageGetDOMStorageItemsResult](d.client, "DOMStorage.getDOMStorageItems", params)
}

func (d DOMStorageDomain) RemoveDOMStorageItem(params DOMStorageRemoveDOMStorageItemParams) (DOMStorageRemoveDOMStorageItemResult, error) {
	return sendCDPCommand[DOMStorageRemoveDOMStorageItemResult](d.client, "DOMStorage.removeDOMStorageItem", params)
}

func (d DOMStorageDomain) SetDOMStorageItem(params DOMStorageSetDOMStorageItemParams) (DOMStorageSetDOMStorageItemResult, error) {
	return sendCDPCommand[DOMStorageSetDOMStorageItemResult](d.client, "DOMStorage.setDOMStorageItem", params)
}

type DebuggerDomain struct {
	client *ModCDPClient
	On     DebuggerEvents
}

func (d DebuggerDomain) ContinueToLocation(params DebuggerContinueToLocationParams) (DebuggerContinueToLocationResult, error) {
	return sendCDPCommand[DebuggerContinueToLocationResult](d.client, "Debugger.continueToLocation", params)
}

func (d DebuggerDomain) Disable(params ...DebuggerDisableParams) (DebuggerDisableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return DebuggerDisableResult{}, err
	}
	return sendCDPCommand[DebuggerDisableResult](d.client, "Debugger.disable", p)
}

func (d DebuggerDomain) Enable(params ...DebuggerEnableParams) (DebuggerEnableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return DebuggerEnableResult{}, err
	}
	return sendCDPCommand[DebuggerEnableResult](d.client, "Debugger.enable", p)
}

func (d DebuggerDomain) EvaluateOnCallFrame(params DebuggerEvaluateOnCallFrameParams) (DebuggerEvaluateOnCallFrameResult, error) {
	return sendCDPCommand[DebuggerEvaluateOnCallFrameResult](d.client, "Debugger.evaluateOnCallFrame", params)
}

func (d DebuggerDomain) GetPossibleBreakpoints(params DebuggerGetPossibleBreakpointsParams) (DebuggerGetPossibleBreakpointsResult, error) {
	return sendCDPCommand[DebuggerGetPossibleBreakpointsResult](d.client, "Debugger.getPossibleBreakpoints", params)
}

func (d DebuggerDomain) GetScriptSource(params DebuggerGetScriptSourceParams) (DebuggerGetScriptSourceResult, error) {
	return sendCDPCommand[DebuggerGetScriptSourceResult](d.client, "Debugger.getScriptSource", params)
}

func (d DebuggerDomain) DisassembleWasmModule(params DebuggerDisassembleWasmModuleParams) (DebuggerDisassembleWasmModuleResult, error) {
	return sendCDPCommand[DebuggerDisassembleWasmModuleResult](d.client, "Debugger.disassembleWasmModule", params)
}

func (d DebuggerDomain) NextWasmDisassemblyChunk(params DebuggerNextWasmDisassemblyChunkParams) (DebuggerNextWasmDisassemblyChunkResult, error) {
	return sendCDPCommand[DebuggerNextWasmDisassemblyChunkResult](d.client, "Debugger.nextWasmDisassemblyChunk", params)
}

func (d DebuggerDomain) GetWasmBytecode(params DebuggerGetWasmBytecodeParams) (DebuggerGetWasmBytecodeResult, error) {
	return sendCDPCommand[DebuggerGetWasmBytecodeResult](d.client, "Debugger.getWasmBytecode", params)
}

func (d DebuggerDomain) GetStackTrace(params DebuggerGetStackTraceParams) (DebuggerGetStackTraceResult, error) {
	return sendCDPCommand[DebuggerGetStackTraceResult](d.client, "Debugger.getStackTrace", params)
}

func (d DebuggerDomain) Pause(params ...DebuggerPauseParams) (DebuggerPauseResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return DebuggerPauseResult{}, err
	}
	return sendCDPCommand[DebuggerPauseResult](d.client, "Debugger.pause", p)
}

func (d DebuggerDomain) PauseOnAsyncCall(params DebuggerPauseOnAsyncCallParams) (DebuggerPauseOnAsyncCallResult, error) {
	return sendCDPCommand[DebuggerPauseOnAsyncCallResult](d.client, "Debugger.pauseOnAsyncCall", params)
}

func (d DebuggerDomain) RemoveBreakpoint(params DebuggerRemoveBreakpointParams) (DebuggerRemoveBreakpointResult, error) {
	return sendCDPCommand[DebuggerRemoveBreakpointResult](d.client, "Debugger.removeBreakpoint", params)
}

func (d DebuggerDomain) RestartFrame(params DebuggerRestartFrameParams) (DebuggerRestartFrameResult, error) {
	return sendCDPCommand[DebuggerRestartFrameResult](d.client, "Debugger.restartFrame", params)
}

func (d DebuggerDomain) Resume(params ...DebuggerResumeParams) (DebuggerResumeResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return DebuggerResumeResult{}, err
	}
	return sendCDPCommand[DebuggerResumeResult](d.client, "Debugger.resume", p)
}

func (d DebuggerDomain) SearchInContent(params DebuggerSearchInContentParams) (DebuggerSearchInContentResult, error) {
	return sendCDPCommand[DebuggerSearchInContentResult](d.client, "Debugger.searchInContent", params)
}

func (d DebuggerDomain) SetAsyncCallStackDepth(params DebuggerSetAsyncCallStackDepthParams) (DebuggerSetAsyncCallStackDepthResult, error) {
	return sendCDPCommand[DebuggerSetAsyncCallStackDepthResult](d.client, "Debugger.setAsyncCallStackDepth", params)
}

func (d DebuggerDomain) SetBlackboxExecutionContexts(params DebuggerSetBlackboxExecutionContextsParams) (DebuggerSetBlackboxExecutionContextsResult, error) {
	return sendCDPCommand[DebuggerSetBlackboxExecutionContextsResult](d.client, "Debugger.setBlackboxExecutionContexts", params)
}

func (d DebuggerDomain) SetBlackboxPatterns(params DebuggerSetBlackboxPatternsParams) (DebuggerSetBlackboxPatternsResult, error) {
	return sendCDPCommand[DebuggerSetBlackboxPatternsResult](d.client, "Debugger.setBlackboxPatterns", params)
}

func (d DebuggerDomain) SetBlackboxedRanges(params DebuggerSetBlackboxedRangesParams) (DebuggerSetBlackboxedRangesResult, error) {
	return sendCDPCommand[DebuggerSetBlackboxedRangesResult](d.client, "Debugger.setBlackboxedRanges", params)
}

func (d DebuggerDomain) SetBreakpoint(params DebuggerSetBreakpointParams) (DebuggerSetBreakpointResult, error) {
	return sendCDPCommand[DebuggerSetBreakpointResult](d.client, "Debugger.setBreakpoint", params)
}

func (d DebuggerDomain) SetInstrumentationBreakpoint(params DebuggerSetInstrumentationBreakpointParams) (DebuggerSetInstrumentationBreakpointResult, error) {
	return sendCDPCommand[DebuggerSetInstrumentationBreakpointResult](d.client, "Debugger.setInstrumentationBreakpoint", params)
}

func (d DebuggerDomain) SetBreakpointByURL(params DebuggerSetBreakpointByURLParams) (DebuggerSetBreakpointByURLResult, error) {
	return sendCDPCommand[DebuggerSetBreakpointByURLResult](d.client, "Debugger.setBreakpointByUrl", params)
}

func (d DebuggerDomain) SetBreakpointOnFunctionCall(params DebuggerSetBreakpointOnFunctionCallParams) (DebuggerSetBreakpointOnFunctionCallResult, error) {
	return sendCDPCommand[DebuggerSetBreakpointOnFunctionCallResult](d.client, "Debugger.setBreakpointOnFunctionCall", params)
}

func (d DebuggerDomain) SetBreakpointsActive(params DebuggerSetBreakpointsActiveParams) (DebuggerSetBreakpointsActiveResult, error) {
	return sendCDPCommand[DebuggerSetBreakpointsActiveResult](d.client, "Debugger.setBreakpointsActive", params)
}

func (d DebuggerDomain) SetPauseOnExceptions(params DebuggerSetPauseOnExceptionsParams) (DebuggerSetPauseOnExceptionsResult, error) {
	return sendCDPCommand[DebuggerSetPauseOnExceptionsResult](d.client, "Debugger.setPauseOnExceptions", params)
}

func (d DebuggerDomain) SetReturnValue(params DebuggerSetReturnValueParams) (DebuggerSetReturnValueResult, error) {
	return sendCDPCommand[DebuggerSetReturnValueResult](d.client, "Debugger.setReturnValue", params)
}

func (d DebuggerDomain) SetScriptSource(params DebuggerSetScriptSourceParams) (DebuggerSetScriptSourceResult, error) {
	return sendCDPCommand[DebuggerSetScriptSourceResult](d.client, "Debugger.setScriptSource", params)
}

func (d DebuggerDomain) SetSkipAllPauses(params DebuggerSetSkipAllPausesParams) (DebuggerSetSkipAllPausesResult, error) {
	return sendCDPCommand[DebuggerSetSkipAllPausesResult](d.client, "Debugger.setSkipAllPauses", params)
}

func (d DebuggerDomain) SetVariableValue(params DebuggerSetVariableValueParams) (DebuggerSetVariableValueResult, error) {
	return sendCDPCommand[DebuggerSetVariableValueResult](d.client, "Debugger.setVariableValue", params)
}

func (d DebuggerDomain) StepInto(params ...DebuggerStepIntoParams) (DebuggerStepIntoResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return DebuggerStepIntoResult{}, err
	}
	return sendCDPCommand[DebuggerStepIntoResult](d.client, "Debugger.stepInto", p)
}

func (d DebuggerDomain) StepOut(params ...DebuggerStepOutParams) (DebuggerStepOutResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return DebuggerStepOutResult{}, err
	}
	return sendCDPCommand[DebuggerStepOutResult](d.client, "Debugger.stepOut", p)
}

func (d DebuggerDomain) StepOver(params ...DebuggerStepOverParams) (DebuggerStepOverResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return DebuggerStepOverResult{}, err
	}
	return sendCDPCommand[DebuggerStepOverResult](d.client, "Debugger.stepOver", p)
}

type DeviceAccessDomain struct {
	client *ModCDPClient
	On     DeviceAccessEvents
}

func (d DeviceAccessDomain) Enable(params ...DeviceAccessEnableParams) (DeviceAccessEnableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return DeviceAccessEnableResult{}, err
	}
	return sendCDPCommand[DeviceAccessEnableResult](d.client, "DeviceAccess.enable", p)
}

func (d DeviceAccessDomain) Disable(params ...DeviceAccessDisableParams) (DeviceAccessDisableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return DeviceAccessDisableResult{}, err
	}
	return sendCDPCommand[DeviceAccessDisableResult](d.client, "DeviceAccess.disable", p)
}

func (d DeviceAccessDomain) SelectPrompt(params DeviceAccessSelectPromptParams) (DeviceAccessSelectPromptResult, error) {
	return sendCDPCommand[DeviceAccessSelectPromptResult](d.client, "DeviceAccess.selectPrompt", params)
}

func (d DeviceAccessDomain) CancelPrompt(params DeviceAccessCancelPromptParams) (DeviceAccessCancelPromptResult, error) {
	return sendCDPCommand[DeviceAccessCancelPromptResult](d.client, "DeviceAccess.cancelPrompt", params)
}

type DeviceOrientationDomain struct{ client *ModCDPClient }

func (d DeviceOrientationDomain) ClearDeviceOrientationOverride(params ...DeviceOrientationClearDeviceOrientationOverrideParams) (DeviceOrientationClearDeviceOrientationOverrideResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return DeviceOrientationClearDeviceOrientationOverrideResult{}, err
	}
	return sendCDPCommand[DeviceOrientationClearDeviceOrientationOverrideResult](d.client, "DeviceOrientation.clearDeviceOrientationOverride", p)
}

func (d DeviceOrientationDomain) SetDeviceOrientationOverride(params DeviceOrientationSetDeviceOrientationOverrideParams) (DeviceOrientationSetDeviceOrientationOverrideResult, error) {
	return sendCDPCommand[DeviceOrientationSetDeviceOrientationOverrideResult](d.client, "DeviceOrientation.setDeviceOrientationOverride", params)
}

type EmulationDomain struct {
	client *ModCDPClient
	On     EmulationEvents
}

func (d EmulationDomain) CanEmulate(params ...EmulationCanEmulateParams) (EmulationCanEmulateResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return EmulationCanEmulateResult{}, err
	}
	return sendCDPCommand[EmulationCanEmulateResult](d.client, "Emulation.canEmulate", p)
}

func (d EmulationDomain) ClearDeviceMetricsOverride(params ...EmulationClearDeviceMetricsOverrideParams) (EmulationClearDeviceMetricsOverrideResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return EmulationClearDeviceMetricsOverrideResult{}, err
	}
	return sendCDPCommand[EmulationClearDeviceMetricsOverrideResult](d.client, "Emulation.clearDeviceMetricsOverride", p)
}

func (d EmulationDomain) ClearGeolocationOverride(params ...EmulationClearGeolocationOverrideParams) (EmulationClearGeolocationOverrideResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return EmulationClearGeolocationOverrideResult{}, err
	}
	return sendCDPCommand[EmulationClearGeolocationOverrideResult](d.client, "Emulation.clearGeolocationOverride", p)
}

func (d EmulationDomain) ResetPageScaleFactor(params ...EmulationResetPageScaleFactorParams) (EmulationResetPageScaleFactorResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return EmulationResetPageScaleFactorResult{}, err
	}
	return sendCDPCommand[EmulationResetPageScaleFactorResult](d.client, "Emulation.resetPageScaleFactor", p)
}

func (d EmulationDomain) SetFocusEmulationEnabled(params EmulationSetFocusEmulationEnabledParams) (EmulationSetFocusEmulationEnabledResult, error) {
	return sendCDPCommand[EmulationSetFocusEmulationEnabledResult](d.client, "Emulation.setFocusEmulationEnabled", params)
}

func (d EmulationDomain) SetAutoDarkModeOverride(params ...EmulationSetAutoDarkModeOverrideParams) (EmulationSetAutoDarkModeOverrideResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return EmulationSetAutoDarkModeOverrideResult{}, err
	}
	return sendCDPCommand[EmulationSetAutoDarkModeOverrideResult](d.client, "Emulation.setAutoDarkModeOverride", p)
}

func (d EmulationDomain) SetCPUThrottlingRate(params EmulationSetCPUThrottlingRateParams) (EmulationSetCPUThrottlingRateResult, error) {
	return sendCDPCommand[EmulationSetCPUThrottlingRateResult](d.client, "Emulation.setCPUThrottlingRate", params)
}

func (d EmulationDomain) SetDefaultBackgroundColorOverride(params ...EmulationSetDefaultBackgroundColorOverrideParams) (EmulationSetDefaultBackgroundColorOverrideResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return EmulationSetDefaultBackgroundColorOverrideResult{}, err
	}
	return sendCDPCommand[EmulationSetDefaultBackgroundColorOverrideResult](d.client, "Emulation.setDefaultBackgroundColorOverride", p)
}

func (d EmulationDomain) SetSafeAreaInsetsOverride(params EmulationSetSafeAreaInsetsOverrideParams) (EmulationSetSafeAreaInsetsOverrideResult, error) {
	return sendCDPCommand[EmulationSetSafeAreaInsetsOverrideResult](d.client, "Emulation.setSafeAreaInsetsOverride", params)
}

func (d EmulationDomain) SetDeviceMetricsOverride(params EmulationSetDeviceMetricsOverrideParams) (EmulationSetDeviceMetricsOverrideResult, error) {
	return sendCDPCommand[EmulationSetDeviceMetricsOverrideResult](d.client, "Emulation.setDeviceMetricsOverride", params)
}

func (d EmulationDomain) SetDevicePostureOverride(params EmulationSetDevicePostureOverrideParams) (EmulationSetDevicePostureOverrideResult, error) {
	return sendCDPCommand[EmulationSetDevicePostureOverrideResult](d.client, "Emulation.setDevicePostureOverride", params)
}

func (d EmulationDomain) ClearDevicePostureOverride(params ...EmulationClearDevicePostureOverrideParams) (EmulationClearDevicePostureOverrideResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return EmulationClearDevicePostureOverrideResult{}, err
	}
	return sendCDPCommand[EmulationClearDevicePostureOverrideResult](d.client, "Emulation.clearDevicePostureOverride", p)
}

func (d EmulationDomain) SetDisplayFeaturesOverride(params EmulationSetDisplayFeaturesOverrideParams) (EmulationSetDisplayFeaturesOverrideResult, error) {
	return sendCDPCommand[EmulationSetDisplayFeaturesOverrideResult](d.client, "Emulation.setDisplayFeaturesOverride", params)
}

func (d EmulationDomain) ClearDisplayFeaturesOverride(params ...EmulationClearDisplayFeaturesOverrideParams) (EmulationClearDisplayFeaturesOverrideResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return EmulationClearDisplayFeaturesOverrideResult{}, err
	}
	return sendCDPCommand[EmulationClearDisplayFeaturesOverrideResult](d.client, "Emulation.clearDisplayFeaturesOverride", p)
}

func (d EmulationDomain) SetScrollbarsHidden(params EmulationSetScrollbarsHiddenParams) (EmulationSetScrollbarsHiddenResult, error) {
	return sendCDPCommand[EmulationSetScrollbarsHiddenResult](d.client, "Emulation.setScrollbarsHidden", params)
}

func (d EmulationDomain) SetDocumentCookieDisabled(params EmulationSetDocumentCookieDisabledParams) (EmulationSetDocumentCookieDisabledResult, error) {
	return sendCDPCommand[EmulationSetDocumentCookieDisabledResult](d.client, "Emulation.setDocumentCookieDisabled", params)
}

func (d EmulationDomain) SetEmitTouchEventsForMouse(params EmulationSetEmitTouchEventsForMouseParams) (EmulationSetEmitTouchEventsForMouseResult, error) {
	return sendCDPCommand[EmulationSetEmitTouchEventsForMouseResult](d.client, "Emulation.setEmitTouchEventsForMouse", params)
}

func (d EmulationDomain) SetEmulatedMedia(params ...EmulationSetEmulatedMediaParams) (EmulationSetEmulatedMediaResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return EmulationSetEmulatedMediaResult{}, err
	}
	return sendCDPCommand[EmulationSetEmulatedMediaResult](d.client, "Emulation.setEmulatedMedia", p)
}

func (d EmulationDomain) SetEmulatedVisionDeficiency(params EmulationSetEmulatedVisionDeficiencyParams) (EmulationSetEmulatedVisionDeficiencyResult, error) {
	return sendCDPCommand[EmulationSetEmulatedVisionDeficiencyResult](d.client, "Emulation.setEmulatedVisionDeficiency", params)
}

func (d EmulationDomain) SetEmulatedOSTextScale(params ...EmulationSetEmulatedOSTextScaleParams) (EmulationSetEmulatedOSTextScaleResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return EmulationSetEmulatedOSTextScaleResult{}, err
	}
	return sendCDPCommand[EmulationSetEmulatedOSTextScaleResult](d.client, "Emulation.setEmulatedOSTextScale", p)
}

func (d EmulationDomain) SetGeolocationOverride(params ...EmulationSetGeolocationOverrideParams) (EmulationSetGeolocationOverrideResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return EmulationSetGeolocationOverrideResult{}, err
	}
	return sendCDPCommand[EmulationSetGeolocationOverrideResult](d.client, "Emulation.setGeolocationOverride", p)
}

func (d EmulationDomain) GetOverriddenSensorInformation(params EmulationGetOverriddenSensorInformationParams) (EmulationGetOverriddenSensorInformationResult, error) {
	return sendCDPCommand[EmulationGetOverriddenSensorInformationResult](d.client, "Emulation.getOverriddenSensorInformation", params)
}

func (d EmulationDomain) SetSensorOverrideEnabled(params EmulationSetSensorOverrideEnabledParams) (EmulationSetSensorOverrideEnabledResult, error) {
	return sendCDPCommand[EmulationSetSensorOverrideEnabledResult](d.client, "Emulation.setSensorOverrideEnabled", params)
}

func (d EmulationDomain) SetSensorOverrideReadings(params EmulationSetSensorOverrideReadingsParams) (EmulationSetSensorOverrideReadingsResult, error) {
	return sendCDPCommand[EmulationSetSensorOverrideReadingsResult](d.client, "Emulation.setSensorOverrideReadings", params)
}

func (d EmulationDomain) SetPressureSourceOverrideEnabled(params EmulationSetPressureSourceOverrideEnabledParams) (EmulationSetPressureSourceOverrideEnabledResult, error) {
	return sendCDPCommand[EmulationSetPressureSourceOverrideEnabledResult](d.client, "Emulation.setPressureSourceOverrideEnabled", params)
}

func (d EmulationDomain) SetPressureStateOverride(params EmulationSetPressureStateOverrideParams) (EmulationSetPressureStateOverrideResult, error) {
	return sendCDPCommand[EmulationSetPressureStateOverrideResult](d.client, "Emulation.setPressureStateOverride", params)
}

func (d EmulationDomain) SetPressureDataOverride(params EmulationSetPressureDataOverrideParams) (EmulationSetPressureDataOverrideResult, error) {
	return sendCDPCommand[EmulationSetPressureDataOverrideResult](d.client, "Emulation.setPressureDataOverride", params)
}

func (d EmulationDomain) SetIdleOverride(params EmulationSetIdleOverrideParams) (EmulationSetIdleOverrideResult, error) {
	return sendCDPCommand[EmulationSetIdleOverrideResult](d.client, "Emulation.setIdleOverride", params)
}

func (d EmulationDomain) ClearIdleOverride(params ...EmulationClearIdleOverrideParams) (EmulationClearIdleOverrideResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return EmulationClearIdleOverrideResult{}, err
	}
	return sendCDPCommand[EmulationClearIdleOverrideResult](d.client, "Emulation.clearIdleOverride", p)
}

func (d EmulationDomain) SetNavigatorOverrides(params EmulationSetNavigatorOverridesParams) (EmulationSetNavigatorOverridesResult, error) {
	return sendCDPCommand[EmulationSetNavigatorOverridesResult](d.client, "Emulation.setNavigatorOverrides", params)
}

func (d EmulationDomain) SetPageScaleFactor(params EmulationSetPageScaleFactorParams) (EmulationSetPageScaleFactorResult, error) {
	return sendCDPCommand[EmulationSetPageScaleFactorResult](d.client, "Emulation.setPageScaleFactor", params)
}

func (d EmulationDomain) SetScriptExecutionDisabled(params EmulationSetScriptExecutionDisabledParams) (EmulationSetScriptExecutionDisabledResult, error) {
	return sendCDPCommand[EmulationSetScriptExecutionDisabledResult](d.client, "Emulation.setScriptExecutionDisabled", params)
}

func (d EmulationDomain) SetTouchEmulationEnabled(params EmulationSetTouchEmulationEnabledParams) (EmulationSetTouchEmulationEnabledResult, error) {
	return sendCDPCommand[EmulationSetTouchEmulationEnabledResult](d.client, "Emulation.setTouchEmulationEnabled", params)
}

func (d EmulationDomain) SetVirtualTimePolicy(params EmulationSetVirtualTimePolicyParams) (EmulationSetVirtualTimePolicyResult, error) {
	return sendCDPCommand[EmulationSetVirtualTimePolicyResult](d.client, "Emulation.setVirtualTimePolicy", params)
}

func (d EmulationDomain) SetLocaleOverride(params ...EmulationSetLocaleOverrideParams) (EmulationSetLocaleOverrideResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return EmulationSetLocaleOverrideResult{}, err
	}
	return sendCDPCommand[EmulationSetLocaleOverrideResult](d.client, "Emulation.setLocaleOverride", p)
}

func (d EmulationDomain) SetTimezoneOverride(params EmulationSetTimezoneOverrideParams) (EmulationSetTimezoneOverrideResult, error) {
	return sendCDPCommand[EmulationSetTimezoneOverrideResult](d.client, "Emulation.setTimezoneOverride", params)
}

func (d EmulationDomain) SetVisibleSize(params EmulationSetVisibleSizeParams) (EmulationSetVisibleSizeResult, error) {
	return sendCDPCommand[EmulationSetVisibleSizeResult](d.client, "Emulation.setVisibleSize", params)
}

func (d EmulationDomain) SetDisabledImageTypes(params EmulationSetDisabledImageTypesParams) (EmulationSetDisabledImageTypesResult, error) {
	return sendCDPCommand[EmulationSetDisabledImageTypesResult](d.client, "Emulation.setDisabledImageTypes", params)
}

func (d EmulationDomain) SetDataSaverOverride(params ...EmulationSetDataSaverOverrideParams) (EmulationSetDataSaverOverrideResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return EmulationSetDataSaverOverrideResult{}, err
	}
	return sendCDPCommand[EmulationSetDataSaverOverrideResult](d.client, "Emulation.setDataSaverOverride", p)
}

func (d EmulationDomain) SetHardwareConcurrencyOverride(params EmulationSetHardwareConcurrencyOverrideParams) (EmulationSetHardwareConcurrencyOverrideResult, error) {
	return sendCDPCommand[EmulationSetHardwareConcurrencyOverrideResult](d.client, "Emulation.setHardwareConcurrencyOverride", params)
}

func (d EmulationDomain) SetUserAgentOverride(params EmulationSetUserAgentOverrideParams) (EmulationSetUserAgentOverrideResult, error) {
	return sendCDPCommand[EmulationSetUserAgentOverrideResult](d.client, "Emulation.setUserAgentOverride", params)
}

func (d EmulationDomain) SetAutomationOverride(params EmulationSetAutomationOverrideParams) (EmulationSetAutomationOverrideResult, error) {
	return sendCDPCommand[EmulationSetAutomationOverrideResult](d.client, "Emulation.setAutomationOverride", params)
}

func (d EmulationDomain) SetSmallViewportHeightDifferenceOverride(params EmulationSetSmallViewportHeightDifferenceOverrideParams) (EmulationSetSmallViewportHeightDifferenceOverrideResult, error) {
	return sendCDPCommand[EmulationSetSmallViewportHeightDifferenceOverrideResult](d.client, "Emulation.setSmallViewportHeightDifferenceOverride", params)
}

func (d EmulationDomain) GetScreenInfos(params ...EmulationGetScreenInfosParams) (EmulationGetScreenInfosResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return EmulationGetScreenInfosResult{}, err
	}
	return sendCDPCommand[EmulationGetScreenInfosResult](d.client, "Emulation.getScreenInfos", p)
}

func (d EmulationDomain) AddScreen(params EmulationAddScreenParams) (EmulationAddScreenResult, error) {
	return sendCDPCommand[EmulationAddScreenResult](d.client, "Emulation.addScreen", params)
}

func (d EmulationDomain) UpdateScreen(params EmulationUpdateScreenParams) (EmulationUpdateScreenResult, error) {
	return sendCDPCommand[EmulationUpdateScreenResult](d.client, "Emulation.updateScreen", params)
}

func (d EmulationDomain) RemoveScreen(params EmulationRemoveScreenParams) (EmulationRemoveScreenResult, error) {
	return sendCDPCommand[EmulationRemoveScreenResult](d.client, "Emulation.removeScreen", params)
}

func (d EmulationDomain) SetPrimaryScreen(params EmulationSetPrimaryScreenParams) (EmulationSetPrimaryScreenResult, error) {
	return sendCDPCommand[EmulationSetPrimaryScreenResult](d.client, "Emulation.setPrimaryScreen", params)
}

type EventBreakpointsDomain struct{ client *ModCDPClient }

func (d EventBreakpointsDomain) SetInstrumentationBreakpoint(params EventBreakpointsSetInstrumentationBreakpointParams) (EventBreakpointsSetInstrumentationBreakpointResult, error) {
	return sendCDPCommand[EventBreakpointsSetInstrumentationBreakpointResult](d.client, "EventBreakpoints.setInstrumentationBreakpoint", params)
}

func (d EventBreakpointsDomain) RemoveInstrumentationBreakpoint(params EventBreakpointsRemoveInstrumentationBreakpointParams) (EventBreakpointsRemoveInstrumentationBreakpointResult, error) {
	return sendCDPCommand[EventBreakpointsRemoveInstrumentationBreakpointResult](d.client, "EventBreakpoints.removeInstrumentationBreakpoint", params)
}

func (d EventBreakpointsDomain) Disable(params ...EventBreakpointsDisableParams) (EventBreakpointsDisableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return EventBreakpointsDisableResult{}, err
	}
	return sendCDPCommand[EventBreakpointsDisableResult](d.client, "EventBreakpoints.disable", p)
}

type ExtensionsDomain struct{ client *ModCDPClient }

func (d ExtensionsDomain) TriggerAction(params ExtensionsTriggerActionParams) (ExtensionsTriggerActionResult, error) {
	return sendCDPCommand[ExtensionsTriggerActionResult](d.client, "Extensions.triggerAction", params)
}

func (d ExtensionsDomain) LoadUnpacked(params ExtensionsLoadUnpackedParams) (ExtensionsLoadUnpackedResult, error) {
	return sendCDPCommand[ExtensionsLoadUnpackedResult](d.client, "Extensions.loadUnpacked", params)
}

func (d ExtensionsDomain) GetExtensions(params ...ExtensionsGetExtensionsParams) (ExtensionsGetExtensionsResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return ExtensionsGetExtensionsResult{}, err
	}
	return sendCDPCommand[ExtensionsGetExtensionsResult](d.client, "Extensions.getExtensions", p)
}

func (d ExtensionsDomain) Uninstall(params ExtensionsUninstallParams) (ExtensionsUninstallResult, error) {
	return sendCDPCommand[ExtensionsUninstallResult](d.client, "Extensions.uninstall", params)
}

func (d ExtensionsDomain) GetStorageItems(params ExtensionsGetStorageItemsParams) (ExtensionsGetStorageItemsResult, error) {
	return sendCDPCommand[ExtensionsGetStorageItemsResult](d.client, "Extensions.getStorageItems", params)
}

func (d ExtensionsDomain) RemoveStorageItems(params ExtensionsRemoveStorageItemsParams) (ExtensionsRemoveStorageItemsResult, error) {
	return sendCDPCommand[ExtensionsRemoveStorageItemsResult](d.client, "Extensions.removeStorageItems", params)
}

func (d ExtensionsDomain) ClearStorageItems(params ExtensionsClearStorageItemsParams) (ExtensionsClearStorageItemsResult, error) {
	return sendCDPCommand[ExtensionsClearStorageItemsResult](d.client, "Extensions.clearStorageItems", params)
}

func (d ExtensionsDomain) SetStorageItems(params ExtensionsSetStorageItemsParams) (ExtensionsSetStorageItemsResult, error) {
	return sendCDPCommand[ExtensionsSetStorageItemsResult](d.client, "Extensions.setStorageItems", params)
}

type FedCmDomain struct {
	client *ModCDPClient
	On     FedCmEvents
}

func (d FedCmDomain) Enable(params ...FedCmEnableParams) (FedCmEnableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return FedCmEnableResult{}, err
	}
	return sendCDPCommand[FedCmEnableResult](d.client, "FedCm.enable", p)
}

func (d FedCmDomain) Disable(params ...FedCmDisableParams) (FedCmDisableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return FedCmDisableResult{}, err
	}
	return sendCDPCommand[FedCmDisableResult](d.client, "FedCm.disable", p)
}

func (d FedCmDomain) SelectAccount(params FedCmSelectAccountParams) (FedCmSelectAccountResult, error) {
	return sendCDPCommand[FedCmSelectAccountResult](d.client, "FedCm.selectAccount", params)
}

func (d FedCmDomain) ClickDialogButton(params FedCmClickDialogButtonParams) (FedCmClickDialogButtonResult, error) {
	return sendCDPCommand[FedCmClickDialogButtonResult](d.client, "FedCm.clickDialogButton", params)
}

func (d FedCmDomain) OpenURL(params FedCmOpenURLParams) (FedCmOpenURLResult, error) {
	return sendCDPCommand[FedCmOpenURLResult](d.client, "FedCm.openUrl", params)
}

func (d FedCmDomain) DismissDialog(params FedCmDismissDialogParams) (FedCmDismissDialogResult, error) {
	return sendCDPCommand[FedCmDismissDialogResult](d.client, "FedCm.dismissDialog", params)
}

func (d FedCmDomain) ResetCooldown(params ...FedCmResetCooldownParams) (FedCmResetCooldownResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return FedCmResetCooldownResult{}, err
	}
	return sendCDPCommand[FedCmResetCooldownResult](d.client, "FedCm.resetCooldown", p)
}

type FetchDomain struct {
	client *ModCDPClient
	On     FetchEvents
}

func (d FetchDomain) Disable(params ...FetchDisableParams) (FetchDisableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return FetchDisableResult{}, err
	}
	return sendCDPCommand[FetchDisableResult](d.client, "Fetch.disable", p)
}

func (d FetchDomain) Enable(params ...FetchEnableParams) (FetchEnableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return FetchEnableResult{}, err
	}
	return sendCDPCommand[FetchEnableResult](d.client, "Fetch.enable", p)
}

func (d FetchDomain) FailRequest(params FetchFailRequestParams) (FetchFailRequestResult, error) {
	return sendCDPCommand[FetchFailRequestResult](d.client, "Fetch.failRequest", params)
}

func (d FetchDomain) FulfillRequest(params FetchFulfillRequestParams) (FetchFulfillRequestResult, error) {
	return sendCDPCommand[FetchFulfillRequestResult](d.client, "Fetch.fulfillRequest", params)
}

func (d FetchDomain) ContinueRequest(params FetchContinueRequestParams) (FetchContinueRequestResult, error) {
	return sendCDPCommand[FetchContinueRequestResult](d.client, "Fetch.continueRequest", params)
}

func (d FetchDomain) ContinueWithAuth(params FetchContinueWithAuthParams) (FetchContinueWithAuthResult, error) {
	return sendCDPCommand[FetchContinueWithAuthResult](d.client, "Fetch.continueWithAuth", params)
}

func (d FetchDomain) ContinueResponse(params FetchContinueResponseParams) (FetchContinueResponseResult, error) {
	return sendCDPCommand[FetchContinueResponseResult](d.client, "Fetch.continueResponse", params)
}

func (d FetchDomain) GetResponseBody(params FetchGetResponseBodyParams) (FetchGetResponseBodyResult, error) {
	return sendCDPCommand[FetchGetResponseBodyResult](d.client, "Fetch.getResponseBody", params)
}

func (d FetchDomain) TakeResponseBodyAsStream(params FetchTakeResponseBodyAsStreamParams) (FetchTakeResponseBodyAsStreamResult, error) {
	return sendCDPCommand[FetchTakeResponseBodyAsStreamResult](d.client, "Fetch.takeResponseBodyAsStream", params)
}

type FileSystemDomain struct{ client *ModCDPClient }

func (d FileSystemDomain) GetDirectory(params FileSystemGetDirectoryParams) (FileSystemGetDirectoryResult, error) {
	return sendCDPCommand[FileSystemGetDirectoryResult](d.client, "FileSystem.getDirectory", params)
}

type HeadlessExperimentalDomain struct{ client *ModCDPClient }

func (d HeadlessExperimentalDomain) BeginFrame(params ...HeadlessExperimentalBeginFrameParams) (HeadlessExperimentalBeginFrameResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return HeadlessExperimentalBeginFrameResult{}, err
	}
	return sendCDPCommand[HeadlessExperimentalBeginFrameResult](d.client, "HeadlessExperimental.beginFrame", p)
}

func (d HeadlessExperimentalDomain) Disable(params ...HeadlessExperimentalDisableParams) (HeadlessExperimentalDisableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return HeadlessExperimentalDisableResult{}, err
	}
	return sendCDPCommand[HeadlessExperimentalDisableResult](d.client, "HeadlessExperimental.disable", p)
}

func (d HeadlessExperimentalDomain) Enable(params ...HeadlessExperimentalEnableParams) (HeadlessExperimentalEnableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return HeadlessExperimentalEnableResult{}, err
	}
	return sendCDPCommand[HeadlessExperimentalEnableResult](d.client, "HeadlessExperimental.enable", p)
}

type HeapProfilerDomain struct {
	client *ModCDPClient
	On     HeapProfilerEvents
}

func (d HeapProfilerDomain) AddInspectedHeapObject(params HeapProfilerAddInspectedHeapObjectParams) (HeapProfilerAddInspectedHeapObjectResult, error) {
	return sendCDPCommand[HeapProfilerAddInspectedHeapObjectResult](d.client, "HeapProfiler.addInspectedHeapObject", params)
}

func (d HeapProfilerDomain) CollectGarbage(params ...HeapProfilerCollectGarbageParams) (HeapProfilerCollectGarbageResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return HeapProfilerCollectGarbageResult{}, err
	}
	return sendCDPCommand[HeapProfilerCollectGarbageResult](d.client, "HeapProfiler.collectGarbage", p)
}

func (d HeapProfilerDomain) Disable(params ...HeapProfilerDisableParams) (HeapProfilerDisableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return HeapProfilerDisableResult{}, err
	}
	return sendCDPCommand[HeapProfilerDisableResult](d.client, "HeapProfiler.disable", p)
}

func (d HeapProfilerDomain) Enable(params ...HeapProfilerEnableParams) (HeapProfilerEnableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return HeapProfilerEnableResult{}, err
	}
	return sendCDPCommand[HeapProfilerEnableResult](d.client, "HeapProfiler.enable", p)
}

func (d HeapProfilerDomain) GetHeapObjectID(params HeapProfilerGetHeapObjectIDParams) (HeapProfilerGetHeapObjectIDResult, error) {
	return sendCDPCommand[HeapProfilerGetHeapObjectIDResult](d.client, "HeapProfiler.getHeapObjectId", params)
}

func (d HeapProfilerDomain) GetObjectByHeapObjectID(params HeapProfilerGetObjectByHeapObjectIDParams) (HeapProfilerGetObjectByHeapObjectIDResult, error) {
	return sendCDPCommand[HeapProfilerGetObjectByHeapObjectIDResult](d.client, "HeapProfiler.getObjectByHeapObjectId", params)
}

func (d HeapProfilerDomain) GetSamplingProfile(params ...HeapProfilerGetSamplingProfileParams) (HeapProfilerGetSamplingProfileResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return HeapProfilerGetSamplingProfileResult{}, err
	}
	return sendCDPCommand[HeapProfilerGetSamplingProfileResult](d.client, "HeapProfiler.getSamplingProfile", p)
}

func (d HeapProfilerDomain) StartSampling(params ...HeapProfilerStartSamplingParams) (HeapProfilerStartSamplingResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return HeapProfilerStartSamplingResult{}, err
	}
	return sendCDPCommand[HeapProfilerStartSamplingResult](d.client, "HeapProfiler.startSampling", p)
}

func (d HeapProfilerDomain) StartTrackingHeapObjects(params ...HeapProfilerStartTrackingHeapObjectsParams) (HeapProfilerStartTrackingHeapObjectsResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return HeapProfilerStartTrackingHeapObjectsResult{}, err
	}
	return sendCDPCommand[HeapProfilerStartTrackingHeapObjectsResult](d.client, "HeapProfiler.startTrackingHeapObjects", p)
}

func (d HeapProfilerDomain) StopSampling(params ...HeapProfilerStopSamplingParams) (HeapProfilerStopSamplingResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return HeapProfilerStopSamplingResult{}, err
	}
	return sendCDPCommand[HeapProfilerStopSamplingResult](d.client, "HeapProfiler.stopSampling", p)
}

func (d HeapProfilerDomain) StopTrackingHeapObjects(params ...HeapProfilerStopTrackingHeapObjectsParams) (HeapProfilerStopTrackingHeapObjectsResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return HeapProfilerStopTrackingHeapObjectsResult{}, err
	}
	return sendCDPCommand[HeapProfilerStopTrackingHeapObjectsResult](d.client, "HeapProfiler.stopTrackingHeapObjects", p)
}

func (d HeapProfilerDomain) TakeHeapSnapshot(params ...HeapProfilerTakeHeapSnapshotParams) (HeapProfilerTakeHeapSnapshotResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return HeapProfilerTakeHeapSnapshotResult{}, err
	}
	return sendCDPCommand[HeapProfilerTakeHeapSnapshotResult](d.client, "HeapProfiler.takeHeapSnapshot", p)
}

type IODomain struct{ client *ModCDPClient }

func (d IODomain) Close(params IOCloseParams) (IOCloseResult, error) {
	return sendCDPCommand[IOCloseResult](d.client, "IO.close", params)
}

func (d IODomain) Read(params IOReadParams) (IOReadResult, error) {
	return sendCDPCommand[IOReadResult](d.client, "IO.read", params)
}

func (d IODomain) ResolveBlob(params IOResolveBlobParams) (IOResolveBlobResult, error) {
	return sendCDPCommand[IOResolveBlobResult](d.client, "IO.resolveBlob", params)
}

type IndexedDBDomain struct{ client *ModCDPClient }

func (d IndexedDBDomain) ClearObjectStore(params IndexedDBClearObjectStoreParams) (IndexedDBClearObjectStoreResult, error) {
	return sendCDPCommand[IndexedDBClearObjectStoreResult](d.client, "IndexedDB.clearObjectStore", params)
}

func (d IndexedDBDomain) DeleteDatabase(params IndexedDBDeleteDatabaseParams) (IndexedDBDeleteDatabaseResult, error) {
	return sendCDPCommand[IndexedDBDeleteDatabaseResult](d.client, "IndexedDB.deleteDatabase", params)
}

func (d IndexedDBDomain) DeleteObjectStoreEntries(params IndexedDBDeleteObjectStoreEntriesParams) (IndexedDBDeleteObjectStoreEntriesResult, error) {
	return sendCDPCommand[IndexedDBDeleteObjectStoreEntriesResult](d.client, "IndexedDB.deleteObjectStoreEntries", params)
}

func (d IndexedDBDomain) Disable(params ...IndexedDBDisableParams) (IndexedDBDisableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return IndexedDBDisableResult{}, err
	}
	return sendCDPCommand[IndexedDBDisableResult](d.client, "IndexedDB.disable", p)
}

func (d IndexedDBDomain) Enable(params ...IndexedDBEnableParams) (IndexedDBEnableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return IndexedDBEnableResult{}, err
	}
	return sendCDPCommand[IndexedDBEnableResult](d.client, "IndexedDB.enable", p)
}

func (d IndexedDBDomain) RequestData(params IndexedDBRequestDataParams) (IndexedDBRequestDataResult, error) {
	return sendCDPCommand[IndexedDBRequestDataResult](d.client, "IndexedDB.requestData", params)
}

func (d IndexedDBDomain) GetMetadata(params IndexedDBGetMetadataParams) (IndexedDBGetMetadataResult, error) {
	return sendCDPCommand[IndexedDBGetMetadataResult](d.client, "IndexedDB.getMetadata", params)
}

func (d IndexedDBDomain) RequestDatabase(params IndexedDBRequestDatabaseParams) (IndexedDBRequestDatabaseResult, error) {
	return sendCDPCommand[IndexedDBRequestDatabaseResult](d.client, "IndexedDB.requestDatabase", params)
}

func (d IndexedDBDomain) RequestDatabaseNames(params ...IndexedDBRequestDatabaseNamesParams) (IndexedDBRequestDatabaseNamesResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return IndexedDBRequestDatabaseNamesResult{}, err
	}
	return sendCDPCommand[IndexedDBRequestDatabaseNamesResult](d.client, "IndexedDB.requestDatabaseNames", p)
}

type InputDomain struct {
	client *ModCDPClient
	On     InputEvents
}

func (d InputDomain) DispatchDragEvent(params InputDispatchDragEventParams) (InputDispatchDragEventResult, error) {
	return sendCDPCommand[InputDispatchDragEventResult](d.client, "Input.dispatchDragEvent", params)
}

func (d InputDomain) DispatchKeyEvent(params InputDispatchKeyEventParams) (InputDispatchKeyEventResult, error) {
	return sendCDPCommand[InputDispatchKeyEventResult](d.client, "Input.dispatchKeyEvent", params)
}

func (d InputDomain) InsertText(params InputInsertTextParams) (InputInsertTextResult, error) {
	return sendCDPCommand[InputInsertTextResult](d.client, "Input.insertText", params)
}

func (d InputDomain) ImeSetComposition(params InputImeSetCompositionParams) (InputImeSetCompositionResult, error) {
	return sendCDPCommand[InputImeSetCompositionResult](d.client, "Input.imeSetComposition", params)
}

func (d InputDomain) DispatchMouseEvent(params InputDispatchMouseEventParams) (InputDispatchMouseEventResult, error) {
	return sendCDPCommand[InputDispatchMouseEventResult](d.client, "Input.dispatchMouseEvent", params)
}

func (d InputDomain) DispatchTouchEvent(params InputDispatchTouchEventParams) (InputDispatchTouchEventResult, error) {
	return sendCDPCommand[InputDispatchTouchEventResult](d.client, "Input.dispatchTouchEvent", params)
}

func (d InputDomain) CancelDragging(params ...InputCancelDraggingParams) (InputCancelDraggingResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return InputCancelDraggingResult{}, err
	}
	return sendCDPCommand[InputCancelDraggingResult](d.client, "Input.cancelDragging", p)
}

func (d InputDomain) EmulateTouchFromMouseEvent(params InputEmulateTouchFromMouseEventParams) (InputEmulateTouchFromMouseEventResult, error) {
	return sendCDPCommand[InputEmulateTouchFromMouseEventResult](d.client, "Input.emulateTouchFromMouseEvent", params)
}

func (d InputDomain) SetIgnoreInputEvents(params InputSetIgnoreInputEventsParams) (InputSetIgnoreInputEventsResult, error) {
	return sendCDPCommand[InputSetIgnoreInputEventsResult](d.client, "Input.setIgnoreInputEvents", params)
}

func (d InputDomain) SetInterceptDrags(params InputSetInterceptDragsParams) (InputSetInterceptDragsResult, error) {
	return sendCDPCommand[InputSetInterceptDragsResult](d.client, "Input.setInterceptDrags", params)
}

func (d InputDomain) SynthesizePinchGesture(params InputSynthesizePinchGestureParams) (InputSynthesizePinchGestureResult, error) {
	return sendCDPCommand[InputSynthesizePinchGestureResult](d.client, "Input.synthesizePinchGesture", params)
}

func (d InputDomain) SynthesizeScrollGesture(params InputSynthesizeScrollGestureParams) (InputSynthesizeScrollGestureResult, error) {
	return sendCDPCommand[InputSynthesizeScrollGestureResult](d.client, "Input.synthesizeScrollGesture", params)
}

func (d InputDomain) SynthesizeTapGesture(params InputSynthesizeTapGestureParams) (InputSynthesizeTapGestureResult, error) {
	return sendCDPCommand[InputSynthesizeTapGestureResult](d.client, "Input.synthesizeTapGesture", params)
}

type InspectorDomain struct {
	client *ModCDPClient
	On     InspectorEvents
}

func (d InspectorDomain) Disable(params ...InspectorDisableParams) (InspectorDisableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return InspectorDisableResult{}, err
	}
	return sendCDPCommand[InspectorDisableResult](d.client, "Inspector.disable", p)
}

func (d InspectorDomain) Enable(params ...InspectorEnableParams) (InspectorEnableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return InspectorEnableResult{}, err
	}
	return sendCDPCommand[InspectorEnableResult](d.client, "Inspector.enable", p)
}

type LayerTreeDomain struct {
	client *ModCDPClient
	On     LayerTreeEvents
}

func (d LayerTreeDomain) CompositingReasons(params LayerTreeCompositingReasonsParams) (LayerTreeCompositingReasonsResult, error) {
	return sendCDPCommand[LayerTreeCompositingReasonsResult](d.client, "LayerTree.compositingReasons", params)
}

func (d LayerTreeDomain) Disable(params ...LayerTreeDisableParams) (LayerTreeDisableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return LayerTreeDisableResult{}, err
	}
	return sendCDPCommand[LayerTreeDisableResult](d.client, "LayerTree.disable", p)
}

func (d LayerTreeDomain) Enable(params ...LayerTreeEnableParams) (LayerTreeEnableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return LayerTreeEnableResult{}, err
	}
	return sendCDPCommand[LayerTreeEnableResult](d.client, "LayerTree.enable", p)
}

func (d LayerTreeDomain) LoadSnapshot(params LayerTreeLoadSnapshotParams) (LayerTreeLoadSnapshotResult, error) {
	return sendCDPCommand[LayerTreeLoadSnapshotResult](d.client, "LayerTree.loadSnapshot", params)
}

func (d LayerTreeDomain) MakeSnapshot(params LayerTreeMakeSnapshotParams) (LayerTreeMakeSnapshotResult, error) {
	return sendCDPCommand[LayerTreeMakeSnapshotResult](d.client, "LayerTree.makeSnapshot", params)
}

func (d LayerTreeDomain) ProfileSnapshot(params LayerTreeProfileSnapshotParams) (LayerTreeProfileSnapshotResult, error) {
	return sendCDPCommand[LayerTreeProfileSnapshotResult](d.client, "LayerTree.profileSnapshot", params)
}

func (d LayerTreeDomain) ReleaseSnapshot(params LayerTreeReleaseSnapshotParams) (LayerTreeReleaseSnapshotResult, error) {
	return sendCDPCommand[LayerTreeReleaseSnapshotResult](d.client, "LayerTree.releaseSnapshot", params)
}

func (d LayerTreeDomain) ReplaySnapshot(params LayerTreeReplaySnapshotParams) (LayerTreeReplaySnapshotResult, error) {
	return sendCDPCommand[LayerTreeReplaySnapshotResult](d.client, "LayerTree.replaySnapshot", params)
}

func (d LayerTreeDomain) SnapshotCommandLog(params LayerTreeSnapshotCommandLogParams) (LayerTreeSnapshotCommandLogResult, error) {
	return sendCDPCommand[LayerTreeSnapshotCommandLogResult](d.client, "LayerTree.snapshotCommandLog", params)
}

type LogDomain struct {
	client *ModCDPClient
	On     LogEvents
}

func (d LogDomain) Clear(params ...LogClearParams) (LogClearResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return LogClearResult{}, err
	}
	return sendCDPCommand[LogClearResult](d.client, "Log.clear", p)
}

func (d LogDomain) Disable(params ...LogDisableParams) (LogDisableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return LogDisableResult{}, err
	}
	return sendCDPCommand[LogDisableResult](d.client, "Log.disable", p)
}

func (d LogDomain) Enable(params ...LogEnableParams) (LogEnableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return LogEnableResult{}, err
	}
	return sendCDPCommand[LogEnableResult](d.client, "Log.enable", p)
}

func (d LogDomain) StartViolationsReport(params LogStartViolationsReportParams) (LogStartViolationsReportResult, error) {
	return sendCDPCommand[LogStartViolationsReportResult](d.client, "Log.startViolationsReport", params)
}

func (d LogDomain) StopViolationsReport(params ...LogStopViolationsReportParams) (LogStopViolationsReportResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return LogStopViolationsReportResult{}, err
	}
	return sendCDPCommand[LogStopViolationsReportResult](d.client, "Log.stopViolationsReport", p)
}

type MediaDomain struct {
	client *ModCDPClient
	On     MediaEvents
}

func (d MediaDomain) Enable(params ...MediaEnableParams) (MediaEnableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return MediaEnableResult{}, err
	}
	return sendCDPCommand[MediaEnableResult](d.client, "Media.enable", p)
}

func (d MediaDomain) Disable(params ...MediaDisableParams) (MediaDisableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return MediaDisableResult{}, err
	}
	return sendCDPCommand[MediaDisableResult](d.client, "Media.disable", p)
}

type MemoryDomain struct{ client *ModCDPClient }

func (d MemoryDomain) GetDOMCounters(params ...MemoryGetDOMCountersParams) (MemoryGetDOMCountersResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return MemoryGetDOMCountersResult{}, err
	}
	return sendCDPCommand[MemoryGetDOMCountersResult](d.client, "Memory.getDOMCounters", p)
}

func (d MemoryDomain) GetDOMCountersForLeakDetection(params ...MemoryGetDOMCountersForLeakDetectionParams) (MemoryGetDOMCountersForLeakDetectionResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return MemoryGetDOMCountersForLeakDetectionResult{}, err
	}
	return sendCDPCommand[MemoryGetDOMCountersForLeakDetectionResult](d.client, "Memory.getDOMCountersForLeakDetection", p)
}

func (d MemoryDomain) PrepareForLeakDetection(params ...MemoryPrepareForLeakDetectionParams) (MemoryPrepareForLeakDetectionResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return MemoryPrepareForLeakDetectionResult{}, err
	}
	return sendCDPCommand[MemoryPrepareForLeakDetectionResult](d.client, "Memory.prepareForLeakDetection", p)
}

func (d MemoryDomain) ForciblyPurgeJavaScriptMemory(params ...MemoryForciblyPurgeJavaScriptMemoryParams) (MemoryForciblyPurgeJavaScriptMemoryResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return MemoryForciblyPurgeJavaScriptMemoryResult{}, err
	}
	return sendCDPCommand[MemoryForciblyPurgeJavaScriptMemoryResult](d.client, "Memory.forciblyPurgeJavaScriptMemory", p)
}

func (d MemoryDomain) SetPressureNotificationsSuppressed(params MemorySetPressureNotificationsSuppressedParams) (MemorySetPressureNotificationsSuppressedResult, error) {
	return sendCDPCommand[MemorySetPressureNotificationsSuppressedResult](d.client, "Memory.setPressureNotificationsSuppressed", params)
}

func (d MemoryDomain) SimulatePressureNotification(params MemorySimulatePressureNotificationParams) (MemorySimulatePressureNotificationResult, error) {
	return sendCDPCommand[MemorySimulatePressureNotificationResult](d.client, "Memory.simulatePressureNotification", params)
}

func (d MemoryDomain) StartSampling(params ...MemoryStartSamplingParams) (MemoryStartSamplingResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return MemoryStartSamplingResult{}, err
	}
	return sendCDPCommand[MemoryStartSamplingResult](d.client, "Memory.startSampling", p)
}

func (d MemoryDomain) StopSampling(params ...MemoryStopSamplingParams) (MemoryStopSamplingResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return MemoryStopSamplingResult{}, err
	}
	return sendCDPCommand[MemoryStopSamplingResult](d.client, "Memory.stopSampling", p)
}

func (d MemoryDomain) GetAllTimeSamplingProfile(params ...MemoryGetAllTimeSamplingProfileParams) (MemoryGetAllTimeSamplingProfileResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return MemoryGetAllTimeSamplingProfileResult{}, err
	}
	return sendCDPCommand[MemoryGetAllTimeSamplingProfileResult](d.client, "Memory.getAllTimeSamplingProfile", p)
}

func (d MemoryDomain) GetBrowserSamplingProfile(params ...MemoryGetBrowserSamplingProfileParams) (MemoryGetBrowserSamplingProfileResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return MemoryGetBrowserSamplingProfileResult{}, err
	}
	return sendCDPCommand[MemoryGetBrowserSamplingProfileResult](d.client, "Memory.getBrowserSamplingProfile", p)
}

func (d MemoryDomain) GetSamplingProfile(params ...MemoryGetSamplingProfileParams) (MemoryGetSamplingProfileResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return MemoryGetSamplingProfileResult{}, err
	}
	return sendCDPCommand[MemoryGetSamplingProfileResult](d.client, "Memory.getSamplingProfile", p)
}

type NetworkDomain struct {
	client *ModCDPClient
	On     NetworkEvents
}

func (d NetworkDomain) SetAcceptedEncodings(params NetworkSetAcceptedEncodingsParams) (NetworkSetAcceptedEncodingsResult, error) {
	return sendCDPCommand[NetworkSetAcceptedEncodingsResult](d.client, "Network.setAcceptedEncodings", params)
}

func (d NetworkDomain) ClearAcceptedEncodingsOverride(params ...NetworkClearAcceptedEncodingsOverrideParams) (NetworkClearAcceptedEncodingsOverrideResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return NetworkClearAcceptedEncodingsOverrideResult{}, err
	}
	return sendCDPCommand[NetworkClearAcceptedEncodingsOverrideResult](d.client, "Network.clearAcceptedEncodingsOverride", p)
}

func (d NetworkDomain) CanClearBrowserCache(params ...NetworkCanClearBrowserCacheParams) (NetworkCanClearBrowserCacheResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return NetworkCanClearBrowserCacheResult{}, err
	}
	return sendCDPCommand[NetworkCanClearBrowserCacheResult](d.client, "Network.canClearBrowserCache", p)
}

func (d NetworkDomain) CanClearBrowserCookies(params ...NetworkCanClearBrowserCookiesParams) (NetworkCanClearBrowserCookiesResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return NetworkCanClearBrowserCookiesResult{}, err
	}
	return sendCDPCommand[NetworkCanClearBrowserCookiesResult](d.client, "Network.canClearBrowserCookies", p)
}

func (d NetworkDomain) CanEmulateNetworkConditions(params ...NetworkCanEmulateNetworkConditionsParams) (NetworkCanEmulateNetworkConditionsResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return NetworkCanEmulateNetworkConditionsResult{}, err
	}
	return sendCDPCommand[NetworkCanEmulateNetworkConditionsResult](d.client, "Network.canEmulateNetworkConditions", p)
}

func (d NetworkDomain) ClearBrowserCache(params ...NetworkClearBrowserCacheParams) (NetworkClearBrowserCacheResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return NetworkClearBrowserCacheResult{}, err
	}
	return sendCDPCommand[NetworkClearBrowserCacheResult](d.client, "Network.clearBrowserCache", p)
}

func (d NetworkDomain) ClearBrowserCookies(params ...NetworkClearBrowserCookiesParams) (NetworkClearBrowserCookiesResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return NetworkClearBrowserCookiesResult{}, err
	}
	return sendCDPCommand[NetworkClearBrowserCookiesResult](d.client, "Network.clearBrowserCookies", p)
}

func (d NetworkDomain) ContinueInterceptedRequest(params NetworkContinueInterceptedRequestParams) (NetworkContinueInterceptedRequestResult, error) {
	return sendCDPCommand[NetworkContinueInterceptedRequestResult](d.client, "Network.continueInterceptedRequest", params)
}

func (d NetworkDomain) DeleteCookies(params NetworkDeleteCookiesParams) (NetworkDeleteCookiesResult, error) {
	return sendCDPCommand[NetworkDeleteCookiesResult](d.client, "Network.deleteCookies", params)
}

func (d NetworkDomain) Disable(params ...NetworkDisableParams) (NetworkDisableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return NetworkDisableResult{}, err
	}
	return sendCDPCommand[NetworkDisableResult](d.client, "Network.disable", p)
}

func (d NetworkDomain) EmulateNetworkConditions(params NetworkEmulateNetworkConditionsParams) (NetworkEmulateNetworkConditionsResult, error) {
	return sendCDPCommand[NetworkEmulateNetworkConditionsResult](d.client, "Network.emulateNetworkConditions", params)
}

func (d NetworkDomain) EmulateNetworkConditionsByRule(params NetworkEmulateNetworkConditionsByRuleParams) (NetworkEmulateNetworkConditionsByRuleResult, error) {
	return sendCDPCommand[NetworkEmulateNetworkConditionsByRuleResult](d.client, "Network.emulateNetworkConditionsByRule", params)
}

func (d NetworkDomain) OverrideNetworkState(params NetworkOverrideNetworkStateParams) (NetworkOverrideNetworkStateResult, error) {
	return sendCDPCommand[NetworkOverrideNetworkStateResult](d.client, "Network.overrideNetworkState", params)
}

func (d NetworkDomain) Enable(params ...NetworkEnableParams) (NetworkEnableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return NetworkEnableResult{}, err
	}
	return sendCDPCommand[NetworkEnableResult](d.client, "Network.enable", p)
}

func (d NetworkDomain) ConfigureDurableMessages(params ...NetworkConfigureDurableMessagesParams) (NetworkConfigureDurableMessagesResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return NetworkConfigureDurableMessagesResult{}, err
	}
	return sendCDPCommand[NetworkConfigureDurableMessagesResult](d.client, "Network.configureDurableMessages", p)
}

func (d NetworkDomain) GetAllCookies(params ...NetworkGetAllCookiesParams) (NetworkGetAllCookiesResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return NetworkGetAllCookiesResult{}, err
	}
	return sendCDPCommand[NetworkGetAllCookiesResult](d.client, "Network.getAllCookies", p)
}

func (d NetworkDomain) GetCertificate(params NetworkGetCertificateParams) (NetworkGetCertificateResult, error) {
	return sendCDPCommand[NetworkGetCertificateResult](d.client, "Network.getCertificate", params)
}

func (d NetworkDomain) GetCookies(params ...NetworkGetCookiesParams) (NetworkGetCookiesResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return NetworkGetCookiesResult{}, err
	}
	return sendCDPCommand[NetworkGetCookiesResult](d.client, "Network.getCookies", p)
}

func (d NetworkDomain) GetResponseBody(params NetworkGetResponseBodyParams) (NetworkGetResponseBodyResult, error) {
	return sendCDPCommand[NetworkGetResponseBodyResult](d.client, "Network.getResponseBody", params)
}

func (d NetworkDomain) GetRequestPostData(params NetworkGetRequestPostDataParams) (NetworkGetRequestPostDataResult, error) {
	return sendCDPCommand[NetworkGetRequestPostDataResult](d.client, "Network.getRequestPostData", params)
}

func (d NetworkDomain) GetResponseBodyForInterception(params NetworkGetResponseBodyForInterceptionParams) (NetworkGetResponseBodyForInterceptionResult, error) {
	return sendCDPCommand[NetworkGetResponseBodyForInterceptionResult](d.client, "Network.getResponseBodyForInterception", params)
}

func (d NetworkDomain) TakeResponseBodyForInterceptionAsStream(params NetworkTakeResponseBodyForInterceptionAsStreamParams) (NetworkTakeResponseBodyForInterceptionAsStreamResult, error) {
	return sendCDPCommand[NetworkTakeResponseBodyForInterceptionAsStreamResult](d.client, "Network.takeResponseBodyForInterceptionAsStream", params)
}

func (d NetworkDomain) ReplayXHR(params NetworkReplayXHRParams) (NetworkReplayXHRResult, error) {
	return sendCDPCommand[NetworkReplayXHRResult](d.client, "Network.replayXHR", params)
}

func (d NetworkDomain) SearchInResponseBody(params NetworkSearchInResponseBodyParams) (NetworkSearchInResponseBodyResult, error) {
	return sendCDPCommand[NetworkSearchInResponseBodyResult](d.client, "Network.searchInResponseBody", params)
}

func (d NetworkDomain) SetBlockedURLs(params ...NetworkSetBlockedURLsParams) (NetworkSetBlockedURLsResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return NetworkSetBlockedURLsResult{}, err
	}
	return sendCDPCommand[NetworkSetBlockedURLsResult](d.client, "Network.setBlockedURLs", p)
}

func (d NetworkDomain) SetBypassServiceWorker(params NetworkSetBypassServiceWorkerParams) (NetworkSetBypassServiceWorkerResult, error) {
	return sendCDPCommand[NetworkSetBypassServiceWorkerResult](d.client, "Network.setBypassServiceWorker", params)
}

func (d NetworkDomain) SetCacheDisabled(params NetworkSetCacheDisabledParams) (NetworkSetCacheDisabledResult, error) {
	return sendCDPCommand[NetworkSetCacheDisabledResult](d.client, "Network.setCacheDisabled", params)
}

func (d NetworkDomain) SetCookie(params NetworkSetCookieParams) (NetworkSetCookieResult, error) {
	return sendCDPCommand[NetworkSetCookieResult](d.client, "Network.setCookie", params)
}

func (d NetworkDomain) SetCookies(params NetworkSetCookiesParams) (NetworkSetCookiesResult, error) {
	return sendCDPCommand[NetworkSetCookiesResult](d.client, "Network.setCookies", params)
}

func (d NetworkDomain) SetExtraHTTPHeaders(params NetworkSetExtraHTTPHeadersParams) (NetworkSetExtraHTTPHeadersResult, error) {
	return sendCDPCommand[NetworkSetExtraHTTPHeadersResult](d.client, "Network.setExtraHTTPHeaders", params)
}

func (d NetworkDomain) SetAttachDebugStack(params NetworkSetAttachDebugStackParams) (NetworkSetAttachDebugStackResult, error) {
	return sendCDPCommand[NetworkSetAttachDebugStackResult](d.client, "Network.setAttachDebugStack", params)
}

func (d NetworkDomain) SetRequestInterception(params NetworkSetRequestInterceptionParams) (NetworkSetRequestInterceptionResult, error) {
	return sendCDPCommand[NetworkSetRequestInterceptionResult](d.client, "Network.setRequestInterception", params)
}

func (d NetworkDomain) SetUserAgentOverride(params NetworkSetUserAgentOverrideParams) (NetworkSetUserAgentOverrideResult, error) {
	return sendCDPCommand[NetworkSetUserAgentOverrideResult](d.client, "Network.setUserAgentOverride", params)
}

func (d NetworkDomain) StreamResourceContent(params NetworkStreamResourceContentParams) (NetworkStreamResourceContentResult, error) {
	return sendCDPCommand[NetworkStreamResourceContentResult](d.client, "Network.streamResourceContent", params)
}

func (d NetworkDomain) GetSecurityIsolationStatus(params ...NetworkGetSecurityIsolationStatusParams) (NetworkGetSecurityIsolationStatusResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return NetworkGetSecurityIsolationStatusResult{}, err
	}
	return sendCDPCommand[NetworkGetSecurityIsolationStatusResult](d.client, "Network.getSecurityIsolationStatus", p)
}

func (d NetworkDomain) EnableReportingAPI(params NetworkEnableReportingAPIParams) (NetworkEnableReportingAPIResult, error) {
	return sendCDPCommand[NetworkEnableReportingAPIResult](d.client, "Network.enableReportingApi", params)
}

func (d NetworkDomain) EnableDeviceBoundSessions(params NetworkEnableDeviceBoundSessionsParams) (NetworkEnableDeviceBoundSessionsResult, error) {
	return sendCDPCommand[NetworkEnableDeviceBoundSessionsResult](d.client, "Network.enableDeviceBoundSessions", params)
}

func (d NetworkDomain) FetchSchemefulSite(params NetworkFetchSchemefulSiteParams) (NetworkFetchSchemefulSiteResult, error) {
	return sendCDPCommand[NetworkFetchSchemefulSiteResult](d.client, "Network.fetchSchemefulSite", params)
}

func (d NetworkDomain) LoadNetworkResource(params NetworkLoadNetworkResourceParams) (NetworkLoadNetworkResourceResult, error) {
	return sendCDPCommand[NetworkLoadNetworkResourceResult](d.client, "Network.loadNetworkResource", params)
}

func (d NetworkDomain) SetCookieControls(params NetworkSetCookieControlsParams) (NetworkSetCookieControlsResult, error) {
	return sendCDPCommand[NetworkSetCookieControlsResult](d.client, "Network.setCookieControls", params)
}

type OverlayDomain struct {
	client *ModCDPClient
	On     OverlayEvents
}

func (d OverlayDomain) Disable(params ...OverlayDisableParams) (OverlayDisableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return OverlayDisableResult{}, err
	}
	return sendCDPCommand[OverlayDisableResult](d.client, "Overlay.disable", p)
}

func (d OverlayDomain) Enable(params ...OverlayEnableParams) (OverlayEnableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return OverlayEnableResult{}, err
	}
	return sendCDPCommand[OverlayEnableResult](d.client, "Overlay.enable", p)
}

func (d OverlayDomain) GetHighlightObjectForTest(params OverlayGetHighlightObjectForTestParams) (OverlayGetHighlightObjectForTestResult, error) {
	return sendCDPCommand[OverlayGetHighlightObjectForTestResult](d.client, "Overlay.getHighlightObjectForTest", params)
}

func (d OverlayDomain) GetGridHighlightObjectsForTest(params OverlayGetGridHighlightObjectsForTestParams) (OverlayGetGridHighlightObjectsForTestResult, error) {
	return sendCDPCommand[OverlayGetGridHighlightObjectsForTestResult](d.client, "Overlay.getGridHighlightObjectsForTest", params)
}

func (d OverlayDomain) GetSourceOrderHighlightObjectForTest(params OverlayGetSourceOrderHighlightObjectForTestParams) (OverlayGetSourceOrderHighlightObjectForTestResult, error) {
	return sendCDPCommand[OverlayGetSourceOrderHighlightObjectForTestResult](d.client, "Overlay.getSourceOrderHighlightObjectForTest", params)
}

func (d OverlayDomain) HideHighlight(params ...OverlayHideHighlightParams) (OverlayHideHighlightResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return OverlayHideHighlightResult{}, err
	}
	return sendCDPCommand[OverlayHideHighlightResult](d.client, "Overlay.hideHighlight", p)
}

func (d OverlayDomain) HighlightFrame(params OverlayHighlightFrameParams) (OverlayHighlightFrameResult, error) {
	return sendCDPCommand[OverlayHighlightFrameResult](d.client, "Overlay.highlightFrame", params)
}

func (d OverlayDomain) HighlightNode(params OverlayHighlightNodeParams) (OverlayHighlightNodeResult, error) {
	return sendCDPCommand[OverlayHighlightNodeResult](d.client, "Overlay.highlightNode", params)
}

func (d OverlayDomain) HighlightQuad(params OverlayHighlightQuadParams) (OverlayHighlightQuadResult, error) {
	return sendCDPCommand[OverlayHighlightQuadResult](d.client, "Overlay.highlightQuad", params)
}

func (d OverlayDomain) HighlightRect(params OverlayHighlightRectParams) (OverlayHighlightRectResult, error) {
	return sendCDPCommand[OverlayHighlightRectResult](d.client, "Overlay.highlightRect", params)
}

func (d OverlayDomain) HighlightSourceOrder(params OverlayHighlightSourceOrderParams) (OverlayHighlightSourceOrderResult, error) {
	return sendCDPCommand[OverlayHighlightSourceOrderResult](d.client, "Overlay.highlightSourceOrder", params)
}

func (d OverlayDomain) SetInspectMode(params OverlaySetInspectModeParams) (OverlaySetInspectModeResult, error) {
	return sendCDPCommand[OverlaySetInspectModeResult](d.client, "Overlay.setInspectMode", params)
}

func (d OverlayDomain) SetShowAdHighlights(params OverlaySetShowAdHighlightsParams) (OverlaySetShowAdHighlightsResult, error) {
	return sendCDPCommand[OverlaySetShowAdHighlightsResult](d.client, "Overlay.setShowAdHighlights", params)
}

func (d OverlayDomain) SetPausedInDebuggerMessage(params ...OverlaySetPausedInDebuggerMessageParams) (OverlaySetPausedInDebuggerMessageResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return OverlaySetPausedInDebuggerMessageResult{}, err
	}
	return sendCDPCommand[OverlaySetPausedInDebuggerMessageResult](d.client, "Overlay.setPausedInDebuggerMessage", p)
}

func (d OverlayDomain) SetShowDebugBorders(params OverlaySetShowDebugBordersParams) (OverlaySetShowDebugBordersResult, error) {
	return sendCDPCommand[OverlaySetShowDebugBordersResult](d.client, "Overlay.setShowDebugBorders", params)
}

func (d OverlayDomain) SetShowFPSCounter(params OverlaySetShowFPSCounterParams) (OverlaySetShowFPSCounterResult, error) {
	return sendCDPCommand[OverlaySetShowFPSCounterResult](d.client, "Overlay.setShowFPSCounter", params)
}

func (d OverlayDomain) SetShowGridOverlays(params OverlaySetShowGridOverlaysParams) (OverlaySetShowGridOverlaysResult, error) {
	return sendCDPCommand[OverlaySetShowGridOverlaysResult](d.client, "Overlay.setShowGridOverlays", params)
}

func (d OverlayDomain) SetShowFlexOverlays(params OverlaySetShowFlexOverlaysParams) (OverlaySetShowFlexOverlaysResult, error) {
	return sendCDPCommand[OverlaySetShowFlexOverlaysResult](d.client, "Overlay.setShowFlexOverlays", params)
}

func (d OverlayDomain) SetShowScrollSnapOverlays(params OverlaySetShowScrollSnapOverlaysParams) (OverlaySetShowScrollSnapOverlaysResult, error) {
	return sendCDPCommand[OverlaySetShowScrollSnapOverlaysResult](d.client, "Overlay.setShowScrollSnapOverlays", params)
}

func (d OverlayDomain) SetShowContainerQueryOverlays(params OverlaySetShowContainerQueryOverlaysParams) (OverlaySetShowContainerQueryOverlaysResult, error) {
	return sendCDPCommand[OverlaySetShowContainerQueryOverlaysResult](d.client, "Overlay.setShowContainerQueryOverlays", params)
}

func (d OverlayDomain) SetShowInspectedElementAnchor(params OverlaySetShowInspectedElementAnchorParams) (OverlaySetShowInspectedElementAnchorResult, error) {
	return sendCDPCommand[OverlaySetShowInspectedElementAnchorResult](d.client, "Overlay.setShowInspectedElementAnchor", params)
}

func (d OverlayDomain) SetShowPaintRects(params OverlaySetShowPaintRectsParams) (OverlaySetShowPaintRectsResult, error) {
	return sendCDPCommand[OverlaySetShowPaintRectsResult](d.client, "Overlay.setShowPaintRects", params)
}

func (d OverlayDomain) SetShowLayoutShiftRegions(params OverlaySetShowLayoutShiftRegionsParams) (OverlaySetShowLayoutShiftRegionsResult, error) {
	return sendCDPCommand[OverlaySetShowLayoutShiftRegionsResult](d.client, "Overlay.setShowLayoutShiftRegions", params)
}

func (d OverlayDomain) SetShowScrollBottleneckRects(params OverlaySetShowScrollBottleneckRectsParams) (OverlaySetShowScrollBottleneckRectsResult, error) {
	return sendCDPCommand[OverlaySetShowScrollBottleneckRectsResult](d.client, "Overlay.setShowScrollBottleneckRects", params)
}

func (d OverlayDomain) SetShowHitTestBorders(params OverlaySetShowHitTestBordersParams) (OverlaySetShowHitTestBordersResult, error) {
	return sendCDPCommand[OverlaySetShowHitTestBordersResult](d.client, "Overlay.setShowHitTestBorders", params)
}

func (d OverlayDomain) SetShowWebVitals(params OverlaySetShowWebVitalsParams) (OverlaySetShowWebVitalsResult, error) {
	return sendCDPCommand[OverlaySetShowWebVitalsResult](d.client, "Overlay.setShowWebVitals", params)
}

func (d OverlayDomain) SetShowViewportSizeOnResize(params OverlaySetShowViewportSizeOnResizeParams) (OverlaySetShowViewportSizeOnResizeResult, error) {
	return sendCDPCommand[OverlaySetShowViewportSizeOnResizeResult](d.client, "Overlay.setShowViewportSizeOnResize", params)
}

func (d OverlayDomain) SetShowHinge(params ...OverlaySetShowHingeParams) (OverlaySetShowHingeResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return OverlaySetShowHingeResult{}, err
	}
	return sendCDPCommand[OverlaySetShowHingeResult](d.client, "Overlay.setShowHinge", p)
}

func (d OverlayDomain) SetShowIsolatedElements(params OverlaySetShowIsolatedElementsParams) (OverlaySetShowIsolatedElementsResult, error) {
	return sendCDPCommand[OverlaySetShowIsolatedElementsResult](d.client, "Overlay.setShowIsolatedElements", params)
}

func (d OverlayDomain) SetShowWindowControlsOverlay(params ...OverlaySetShowWindowControlsOverlayParams) (OverlaySetShowWindowControlsOverlayResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return OverlaySetShowWindowControlsOverlayResult{}, err
	}
	return sendCDPCommand[OverlaySetShowWindowControlsOverlayResult](d.client, "Overlay.setShowWindowControlsOverlay", p)
}

type PWADomain struct{ client *ModCDPClient }

func (d PWADomain) GetOsAppState(params PWAGetOsAppStateParams) (PWAGetOsAppStateResult, error) {
	return sendCDPCommand[PWAGetOsAppStateResult](d.client, "PWA.getOsAppState", params)
}

func (d PWADomain) Install(params PWAInstallParams) (PWAInstallResult, error) {
	return sendCDPCommand[PWAInstallResult](d.client, "PWA.install", params)
}

func (d PWADomain) Uninstall(params PWAUninstallParams) (PWAUninstallResult, error) {
	return sendCDPCommand[PWAUninstallResult](d.client, "PWA.uninstall", params)
}

func (d PWADomain) Launch(params PWALaunchParams) (PWALaunchResult, error) {
	return sendCDPCommand[PWALaunchResult](d.client, "PWA.launch", params)
}

func (d PWADomain) LaunchFilesInApp(params PWALaunchFilesInAppParams) (PWALaunchFilesInAppResult, error) {
	return sendCDPCommand[PWALaunchFilesInAppResult](d.client, "PWA.launchFilesInApp", params)
}

func (d PWADomain) OpenCurrentPageInApp(params PWAOpenCurrentPageInAppParams) (PWAOpenCurrentPageInAppResult, error) {
	return sendCDPCommand[PWAOpenCurrentPageInAppResult](d.client, "PWA.openCurrentPageInApp", params)
}

func (d PWADomain) ChangeAppUserSettings(params PWAChangeAppUserSettingsParams) (PWAChangeAppUserSettingsResult, error) {
	return sendCDPCommand[PWAChangeAppUserSettingsResult](d.client, "PWA.changeAppUserSettings", params)
}

type PageDomain struct {
	client *ModCDPClient
	On     PageEvents
}

func (d PageDomain) AddScriptToEvaluateOnLoad(params PageAddScriptToEvaluateOnLoadParams) (PageAddScriptToEvaluateOnLoadResult, error) {
	return sendCDPCommand[PageAddScriptToEvaluateOnLoadResult](d.client, "Page.addScriptToEvaluateOnLoad", params)
}

func (d PageDomain) AddScriptToEvaluateOnNewDocument(params PageAddScriptToEvaluateOnNewDocumentParams) (PageAddScriptToEvaluateOnNewDocumentResult, error) {
	return sendCDPCommand[PageAddScriptToEvaluateOnNewDocumentResult](d.client, "Page.addScriptToEvaluateOnNewDocument", params)
}

func (d PageDomain) BringToFront(params ...PageBringToFrontParams) (PageBringToFrontResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return PageBringToFrontResult{}, err
	}
	return sendCDPCommand[PageBringToFrontResult](d.client, "Page.bringToFront", p)
}

func (d PageDomain) CaptureScreenshot(params ...PageCaptureScreenshotParams) (PageCaptureScreenshotResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return PageCaptureScreenshotResult{}, err
	}
	return sendCDPCommand[PageCaptureScreenshotResult](d.client, "Page.captureScreenshot", p)
}

func (d PageDomain) CaptureSnapshot(params ...PageCaptureSnapshotParams) (PageCaptureSnapshotResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return PageCaptureSnapshotResult{}, err
	}
	return sendCDPCommand[PageCaptureSnapshotResult](d.client, "Page.captureSnapshot", p)
}

func (d PageDomain) ClearDeviceMetricsOverride(params ...PageClearDeviceMetricsOverrideParams) (PageClearDeviceMetricsOverrideResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return PageClearDeviceMetricsOverrideResult{}, err
	}
	return sendCDPCommand[PageClearDeviceMetricsOverrideResult](d.client, "Page.clearDeviceMetricsOverride", p)
}

func (d PageDomain) ClearDeviceOrientationOverride(params ...PageClearDeviceOrientationOverrideParams) (PageClearDeviceOrientationOverrideResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return PageClearDeviceOrientationOverrideResult{}, err
	}
	return sendCDPCommand[PageClearDeviceOrientationOverrideResult](d.client, "Page.clearDeviceOrientationOverride", p)
}

func (d PageDomain) ClearGeolocationOverride(params ...PageClearGeolocationOverrideParams) (PageClearGeolocationOverrideResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return PageClearGeolocationOverrideResult{}, err
	}
	return sendCDPCommand[PageClearGeolocationOverrideResult](d.client, "Page.clearGeolocationOverride", p)
}

func (d PageDomain) CreateIsolatedWorld(params PageCreateIsolatedWorldParams) (PageCreateIsolatedWorldResult, error) {
	return sendCDPCommand[PageCreateIsolatedWorldResult](d.client, "Page.createIsolatedWorld", params)
}

func (d PageDomain) DeleteCookie(params PageDeleteCookieParams) (PageDeleteCookieResult, error) {
	return sendCDPCommand[PageDeleteCookieResult](d.client, "Page.deleteCookie", params)
}

func (d PageDomain) Disable(params ...PageDisableParams) (PageDisableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return PageDisableResult{}, err
	}
	return sendCDPCommand[PageDisableResult](d.client, "Page.disable", p)
}

func (d PageDomain) Enable(params ...PageEnableParams) (PageEnableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return PageEnableResult{}, err
	}
	return sendCDPCommand[PageEnableResult](d.client, "Page.enable", p)
}

func (d PageDomain) GetAppManifest(params ...PageGetAppManifestParams) (PageGetAppManifestResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return PageGetAppManifestResult{}, err
	}
	return sendCDPCommand[PageGetAppManifestResult](d.client, "Page.getAppManifest", p)
}

func (d PageDomain) GetInstallabilityErrors(params ...PageGetInstallabilityErrorsParams) (PageGetInstallabilityErrorsResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return PageGetInstallabilityErrorsResult{}, err
	}
	return sendCDPCommand[PageGetInstallabilityErrorsResult](d.client, "Page.getInstallabilityErrors", p)
}

func (d PageDomain) GetManifestIcons(params ...PageGetManifestIconsParams) (PageGetManifestIconsResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return PageGetManifestIconsResult{}, err
	}
	return sendCDPCommand[PageGetManifestIconsResult](d.client, "Page.getManifestIcons", p)
}

func (d PageDomain) GetAppID(params ...PageGetAppIDParams) (PageGetAppIDResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return PageGetAppIDResult{}, err
	}
	return sendCDPCommand[PageGetAppIDResult](d.client, "Page.getAppId", p)
}

func (d PageDomain) GetAdScriptAncestry(params PageGetAdScriptAncestryParams) (PageGetAdScriptAncestryResult, error) {
	return sendCDPCommand[PageGetAdScriptAncestryResult](d.client, "Page.getAdScriptAncestry", params)
}

func (d PageDomain) GetFrameTree(params ...PageGetFrameTreeParams) (PageGetFrameTreeResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return PageGetFrameTreeResult{}, err
	}
	return sendCDPCommand[PageGetFrameTreeResult](d.client, "Page.getFrameTree", p)
}

func (d PageDomain) GetLayoutMetrics(params ...PageGetLayoutMetricsParams) (PageGetLayoutMetricsResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return PageGetLayoutMetricsResult{}, err
	}
	return sendCDPCommand[PageGetLayoutMetricsResult](d.client, "Page.getLayoutMetrics", p)
}

func (d PageDomain) GetNavigationHistory(params ...PageGetNavigationHistoryParams) (PageGetNavigationHistoryResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return PageGetNavigationHistoryResult{}, err
	}
	return sendCDPCommand[PageGetNavigationHistoryResult](d.client, "Page.getNavigationHistory", p)
}

func (d PageDomain) ResetNavigationHistory(params ...PageResetNavigationHistoryParams) (PageResetNavigationHistoryResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return PageResetNavigationHistoryResult{}, err
	}
	return sendCDPCommand[PageResetNavigationHistoryResult](d.client, "Page.resetNavigationHistory", p)
}

func (d PageDomain) GetResourceContent(params PageGetResourceContentParams) (PageGetResourceContentResult, error) {
	return sendCDPCommand[PageGetResourceContentResult](d.client, "Page.getResourceContent", params)
}

func (d PageDomain) GetResourceTree(params ...PageGetResourceTreeParams) (PageGetResourceTreeResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return PageGetResourceTreeResult{}, err
	}
	return sendCDPCommand[PageGetResourceTreeResult](d.client, "Page.getResourceTree", p)
}

func (d PageDomain) HandleJavaScriptDialog(params PageHandleJavaScriptDialogParams) (PageHandleJavaScriptDialogResult, error) {
	return sendCDPCommand[PageHandleJavaScriptDialogResult](d.client, "Page.handleJavaScriptDialog", params)
}

func (d PageDomain) Navigate(params PageNavigateParams) (PageNavigateResult, error) {
	return sendCDPCommand[PageNavigateResult](d.client, "Page.navigate", params)
}

func (d PageDomain) NavigateToHistoryEntry(params PageNavigateToHistoryEntryParams) (PageNavigateToHistoryEntryResult, error) {
	return sendCDPCommand[PageNavigateToHistoryEntryResult](d.client, "Page.navigateToHistoryEntry", params)
}

func (d PageDomain) PrintToPDF(params ...PagePrintToPDFParams) (PagePrintToPDFResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return PagePrintToPDFResult{}, err
	}
	return sendCDPCommand[PagePrintToPDFResult](d.client, "Page.printToPDF", p)
}

func (d PageDomain) Reload(params ...PageReloadParams) (PageReloadResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return PageReloadResult{}, err
	}
	return sendCDPCommand[PageReloadResult](d.client, "Page.reload", p)
}

func (d PageDomain) RemoveScriptToEvaluateOnLoad(params PageRemoveScriptToEvaluateOnLoadParams) (PageRemoveScriptToEvaluateOnLoadResult, error) {
	return sendCDPCommand[PageRemoveScriptToEvaluateOnLoadResult](d.client, "Page.removeScriptToEvaluateOnLoad", params)
}

func (d PageDomain) RemoveScriptToEvaluateOnNewDocument(params PageRemoveScriptToEvaluateOnNewDocumentParams) (PageRemoveScriptToEvaluateOnNewDocumentResult, error) {
	return sendCDPCommand[PageRemoveScriptToEvaluateOnNewDocumentResult](d.client, "Page.removeScriptToEvaluateOnNewDocument", params)
}

func (d PageDomain) ScreencastFrameAck(params PageScreencastFrameAckParams) (PageScreencastFrameAckResult, error) {
	return sendCDPCommand[PageScreencastFrameAckResult](d.client, "Page.screencastFrameAck", params)
}

func (d PageDomain) SearchInResource(params PageSearchInResourceParams) (PageSearchInResourceResult, error) {
	return sendCDPCommand[PageSearchInResourceResult](d.client, "Page.searchInResource", params)
}

func (d PageDomain) SetAdBlockingEnabled(params PageSetAdBlockingEnabledParams) (PageSetAdBlockingEnabledResult, error) {
	return sendCDPCommand[PageSetAdBlockingEnabledResult](d.client, "Page.setAdBlockingEnabled", params)
}

func (d PageDomain) SetBypassCSP(params PageSetBypassCSPParams) (PageSetBypassCSPResult, error) {
	return sendCDPCommand[PageSetBypassCSPResult](d.client, "Page.setBypassCSP", params)
}

func (d PageDomain) GetPermissionsPolicyState(params PageGetPermissionsPolicyStateParams) (PageGetPermissionsPolicyStateResult, error) {
	return sendCDPCommand[PageGetPermissionsPolicyStateResult](d.client, "Page.getPermissionsPolicyState", params)
}

func (d PageDomain) GetOriginTrials(params PageGetOriginTrialsParams) (PageGetOriginTrialsResult, error) {
	return sendCDPCommand[PageGetOriginTrialsResult](d.client, "Page.getOriginTrials", params)
}

func (d PageDomain) SetDeviceMetricsOverride(params PageSetDeviceMetricsOverrideParams) (PageSetDeviceMetricsOverrideResult, error) {
	return sendCDPCommand[PageSetDeviceMetricsOverrideResult](d.client, "Page.setDeviceMetricsOverride", params)
}

func (d PageDomain) SetDeviceOrientationOverride(params PageSetDeviceOrientationOverrideParams) (PageSetDeviceOrientationOverrideResult, error) {
	return sendCDPCommand[PageSetDeviceOrientationOverrideResult](d.client, "Page.setDeviceOrientationOverride", params)
}

func (d PageDomain) SetFontFamilies(params PageSetFontFamiliesParams) (PageSetFontFamiliesResult, error) {
	return sendCDPCommand[PageSetFontFamiliesResult](d.client, "Page.setFontFamilies", params)
}

func (d PageDomain) SetFontSizes(params PageSetFontSizesParams) (PageSetFontSizesResult, error) {
	return sendCDPCommand[PageSetFontSizesResult](d.client, "Page.setFontSizes", params)
}

func (d PageDomain) SetDocumentContent(params PageSetDocumentContentParams) (PageSetDocumentContentResult, error) {
	return sendCDPCommand[PageSetDocumentContentResult](d.client, "Page.setDocumentContent", params)
}

func (d PageDomain) SetDownloadBehavior(params PageSetDownloadBehaviorParams) (PageSetDownloadBehaviorResult, error) {
	return sendCDPCommand[PageSetDownloadBehaviorResult](d.client, "Page.setDownloadBehavior", params)
}

func (d PageDomain) SetGeolocationOverride(params ...PageSetGeolocationOverrideParams) (PageSetGeolocationOverrideResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return PageSetGeolocationOverrideResult{}, err
	}
	return sendCDPCommand[PageSetGeolocationOverrideResult](d.client, "Page.setGeolocationOverride", p)
}

func (d PageDomain) SetLifecycleEventsEnabled(params PageSetLifecycleEventsEnabledParams) (PageSetLifecycleEventsEnabledResult, error) {
	return sendCDPCommand[PageSetLifecycleEventsEnabledResult](d.client, "Page.setLifecycleEventsEnabled", params)
}

func (d PageDomain) SetTouchEmulationEnabled(params PageSetTouchEmulationEnabledParams) (PageSetTouchEmulationEnabledResult, error) {
	return sendCDPCommand[PageSetTouchEmulationEnabledResult](d.client, "Page.setTouchEmulationEnabled", params)
}

func (d PageDomain) StartScreencast(params ...PageStartScreencastParams) (PageStartScreencastResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return PageStartScreencastResult{}, err
	}
	return sendCDPCommand[PageStartScreencastResult](d.client, "Page.startScreencast", p)
}

func (d PageDomain) StopLoading(params ...PageStopLoadingParams) (PageStopLoadingResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return PageStopLoadingResult{}, err
	}
	return sendCDPCommand[PageStopLoadingResult](d.client, "Page.stopLoading", p)
}

func (d PageDomain) Crash(params ...PageCrashParams) (PageCrashResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return PageCrashResult{}, err
	}
	return sendCDPCommand[PageCrashResult](d.client, "Page.crash", p)
}

func (d PageDomain) Close(params ...PageCloseParams) (PageCloseResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return PageCloseResult{}, err
	}
	return sendCDPCommand[PageCloseResult](d.client, "Page.close", p)
}

func (d PageDomain) SetWebLifecycleState(params PageSetWebLifecycleStateParams) (PageSetWebLifecycleStateResult, error) {
	return sendCDPCommand[PageSetWebLifecycleStateResult](d.client, "Page.setWebLifecycleState", params)
}

func (d PageDomain) StopScreencast(params ...PageStopScreencastParams) (PageStopScreencastResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return PageStopScreencastResult{}, err
	}
	return sendCDPCommand[PageStopScreencastResult](d.client, "Page.stopScreencast", p)
}

func (d PageDomain) ProduceCompilationCache(params PageProduceCompilationCacheParams) (PageProduceCompilationCacheResult, error) {
	return sendCDPCommand[PageProduceCompilationCacheResult](d.client, "Page.produceCompilationCache", params)
}

func (d PageDomain) AddCompilationCache(params PageAddCompilationCacheParams) (PageAddCompilationCacheResult, error) {
	return sendCDPCommand[PageAddCompilationCacheResult](d.client, "Page.addCompilationCache", params)
}

func (d PageDomain) ClearCompilationCache(params ...PageClearCompilationCacheParams) (PageClearCompilationCacheResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return PageClearCompilationCacheResult{}, err
	}
	return sendCDPCommand[PageClearCompilationCacheResult](d.client, "Page.clearCompilationCache", p)
}

func (d PageDomain) SetSPCTransactionMode(params PageSetSPCTransactionModeParams) (PageSetSPCTransactionModeResult, error) {
	return sendCDPCommand[PageSetSPCTransactionModeResult](d.client, "Page.setSPCTransactionMode", params)
}

func (d PageDomain) SetRPHRegistrationMode(params PageSetRPHRegistrationModeParams) (PageSetRPHRegistrationModeResult, error) {
	return sendCDPCommand[PageSetRPHRegistrationModeResult](d.client, "Page.setRPHRegistrationMode", params)
}

func (d PageDomain) GenerateTestReport(params PageGenerateTestReportParams) (PageGenerateTestReportResult, error) {
	return sendCDPCommand[PageGenerateTestReportResult](d.client, "Page.generateTestReport", params)
}

func (d PageDomain) WaitForDebugger(params ...PageWaitForDebuggerParams) (PageWaitForDebuggerResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return PageWaitForDebuggerResult{}, err
	}
	return sendCDPCommand[PageWaitForDebuggerResult](d.client, "Page.waitForDebugger", p)
}

func (d PageDomain) SetInterceptFileChooserDialog(params PageSetInterceptFileChooserDialogParams) (PageSetInterceptFileChooserDialogResult, error) {
	return sendCDPCommand[PageSetInterceptFileChooserDialogResult](d.client, "Page.setInterceptFileChooserDialog", params)
}

func (d PageDomain) SetPrerenderingAllowed(params PageSetPrerenderingAllowedParams) (PageSetPrerenderingAllowedResult, error) {
	return sendCDPCommand[PageSetPrerenderingAllowedResult](d.client, "Page.setPrerenderingAllowed", params)
}

func (d PageDomain) GetAnnotatedPageContent(params ...PageGetAnnotatedPageContentParams) (PageGetAnnotatedPageContentResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return PageGetAnnotatedPageContentResult{}, err
	}
	return sendCDPCommand[PageGetAnnotatedPageContentResult](d.client, "Page.getAnnotatedPageContent", p)
}

type PerformanceDomain struct {
	client *ModCDPClient
	On     PerformanceEvents
}

func (d PerformanceDomain) Disable(params ...PerformanceDisableParams) (PerformanceDisableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return PerformanceDisableResult{}, err
	}
	return sendCDPCommand[PerformanceDisableResult](d.client, "Performance.disable", p)
}

func (d PerformanceDomain) Enable(params ...PerformanceEnableParams) (PerformanceEnableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return PerformanceEnableResult{}, err
	}
	return sendCDPCommand[PerformanceEnableResult](d.client, "Performance.enable", p)
}

func (d PerformanceDomain) SetTimeDomain(params PerformanceSetTimeDomainParams) (PerformanceSetTimeDomainResult, error) {
	return sendCDPCommand[PerformanceSetTimeDomainResult](d.client, "Performance.setTimeDomain", params)
}

func (d PerformanceDomain) GetMetrics(params ...PerformanceGetMetricsParams) (PerformanceGetMetricsResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return PerformanceGetMetricsResult{}, err
	}
	return sendCDPCommand[PerformanceGetMetricsResult](d.client, "Performance.getMetrics", p)
}

type PerformanceTimelineDomain struct {
	client *ModCDPClient
	On     PerformanceTimelineEvents
}

func (d PerformanceTimelineDomain) Enable(params PerformanceTimelineEnableParams) (PerformanceTimelineEnableResult, error) {
	return sendCDPCommand[PerformanceTimelineEnableResult](d.client, "PerformanceTimeline.enable", params)
}

type PreloadDomain struct {
	client *ModCDPClient
	On     PreloadEvents
}

func (d PreloadDomain) Enable(params ...PreloadEnableParams) (PreloadEnableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return PreloadEnableResult{}, err
	}
	return sendCDPCommand[PreloadEnableResult](d.client, "Preload.enable", p)
}

func (d PreloadDomain) Disable(params ...PreloadDisableParams) (PreloadDisableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return PreloadDisableResult{}, err
	}
	return sendCDPCommand[PreloadDisableResult](d.client, "Preload.disable", p)
}

type ProfilerDomain struct {
	client *ModCDPClient
	On     ProfilerEvents
}

func (d ProfilerDomain) Disable(params ...ProfilerDisableParams) (ProfilerDisableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return ProfilerDisableResult{}, err
	}
	return sendCDPCommand[ProfilerDisableResult](d.client, "Profiler.disable", p)
}

func (d ProfilerDomain) Enable(params ...ProfilerEnableParams) (ProfilerEnableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return ProfilerEnableResult{}, err
	}
	return sendCDPCommand[ProfilerEnableResult](d.client, "Profiler.enable", p)
}

func (d ProfilerDomain) GetBestEffortCoverage(params ...ProfilerGetBestEffortCoverageParams) (ProfilerGetBestEffortCoverageResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return ProfilerGetBestEffortCoverageResult{}, err
	}
	return sendCDPCommand[ProfilerGetBestEffortCoverageResult](d.client, "Profiler.getBestEffortCoverage", p)
}

func (d ProfilerDomain) SetSamplingInterval(params ProfilerSetSamplingIntervalParams) (ProfilerSetSamplingIntervalResult, error) {
	return sendCDPCommand[ProfilerSetSamplingIntervalResult](d.client, "Profiler.setSamplingInterval", params)
}

func (d ProfilerDomain) Start(params ...ProfilerStartParams) (ProfilerStartResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return ProfilerStartResult{}, err
	}
	return sendCDPCommand[ProfilerStartResult](d.client, "Profiler.start", p)
}

func (d ProfilerDomain) StartPreciseCoverage(params ...ProfilerStartPreciseCoverageParams) (ProfilerStartPreciseCoverageResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return ProfilerStartPreciseCoverageResult{}, err
	}
	return sendCDPCommand[ProfilerStartPreciseCoverageResult](d.client, "Profiler.startPreciseCoverage", p)
}

func (d ProfilerDomain) Stop(params ...ProfilerStopParams) (ProfilerStopResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return ProfilerStopResult{}, err
	}
	return sendCDPCommand[ProfilerStopResult](d.client, "Profiler.stop", p)
}

func (d ProfilerDomain) StopPreciseCoverage(params ...ProfilerStopPreciseCoverageParams) (ProfilerStopPreciseCoverageResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return ProfilerStopPreciseCoverageResult{}, err
	}
	return sendCDPCommand[ProfilerStopPreciseCoverageResult](d.client, "Profiler.stopPreciseCoverage", p)
}

func (d ProfilerDomain) TakePreciseCoverage(params ...ProfilerTakePreciseCoverageParams) (ProfilerTakePreciseCoverageResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return ProfilerTakePreciseCoverageResult{}, err
	}
	return sendCDPCommand[ProfilerTakePreciseCoverageResult](d.client, "Profiler.takePreciseCoverage", p)
}

type RuntimeDomain struct {
	client *ModCDPClient
	On     RuntimeEvents
}

func (d RuntimeDomain) AwaitPromise(params RuntimeAwaitPromiseParams) (RuntimeAwaitPromiseResult, error) {
	return sendCDPCommand[RuntimeAwaitPromiseResult](d.client, "Runtime.awaitPromise", params)
}

func (d RuntimeDomain) CallFunctionOn(params RuntimeCallFunctionOnParams) (RuntimeCallFunctionOnResult, error) {
	return sendCDPCommand[RuntimeCallFunctionOnResult](d.client, "Runtime.callFunctionOn", params)
}

func (d RuntimeDomain) CompileScript(params RuntimeCompileScriptParams) (RuntimeCompileScriptResult, error) {
	return sendCDPCommand[RuntimeCompileScriptResult](d.client, "Runtime.compileScript", params)
}

func (d RuntimeDomain) Disable(params ...RuntimeDisableParams) (RuntimeDisableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return RuntimeDisableResult{}, err
	}
	return sendCDPCommand[RuntimeDisableResult](d.client, "Runtime.disable", p)
}

func (d RuntimeDomain) DiscardConsoleEntries(params ...RuntimeDiscardConsoleEntriesParams) (RuntimeDiscardConsoleEntriesResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return RuntimeDiscardConsoleEntriesResult{}, err
	}
	return sendCDPCommand[RuntimeDiscardConsoleEntriesResult](d.client, "Runtime.discardConsoleEntries", p)
}

func (d RuntimeDomain) Enable(params ...RuntimeEnableParams) (RuntimeEnableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return RuntimeEnableResult{}, err
	}
	return sendCDPCommand[RuntimeEnableResult](d.client, "Runtime.enable", p)
}

func (d RuntimeDomain) Evaluate(params RuntimeEvaluateParams) (RuntimeEvaluateResult, error) {
	return sendCDPCommand[RuntimeEvaluateResult](d.client, "Runtime.evaluate", params)
}

func (d RuntimeDomain) GetIsolateID(params ...RuntimeGetIsolateIDParams) (RuntimeGetIsolateIDResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return RuntimeGetIsolateIDResult{}, err
	}
	return sendCDPCommand[RuntimeGetIsolateIDResult](d.client, "Runtime.getIsolateId", p)
}

func (d RuntimeDomain) GetHeapUsage(params ...RuntimeGetHeapUsageParams) (RuntimeGetHeapUsageResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return RuntimeGetHeapUsageResult{}, err
	}
	return sendCDPCommand[RuntimeGetHeapUsageResult](d.client, "Runtime.getHeapUsage", p)
}

func (d RuntimeDomain) GetProperties(params RuntimeGetPropertiesParams) (RuntimeGetPropertiesResult, error) {
	return sendCDPCommand[RuntimeGetPropertiesResult](d.client, "Runtime.getProperties", params)
}

func (d RuntimeDomain) GlobalLexicalScopeNames(params ...RuntimeGlobalLexicalScopeNamesParams) (RuntimeGlobalLexicalScopeNamesResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return RuntimeGlobalLexicalScopeNamesResult{}, err
	}
	return sendCDPCommand[RuntimeGlobalLexicalScopeNamesResult](d.client, "Runtime.globalLexicalScopeNames", p)
}

func (d RuntimeDomain) QueryObjects(params RuntimeQueryObjectsParams) (RuntimeQueryObjectsResult, error) {
	return sendCDPCommand[RuntimeQueryObjectsResult](d.client, "Runtime.queryObjects", params)
}

func (d RuntimeDomain) ReleaseObject(params RuntimeReleaseObjectParams) (RuntimeReleaseObjectResult, error) {
	return sendCDPCommand[RuntimeReleaseObjectResult](d.client, "Runtime.releaseObject", params)
}

func (d RuntimeDomain) ReleaseObjectGroup(params RuntimeReleaseObjectGroupParams) (RuntimeReleaseObjectGroupResult, error) {
	return sendCDPCommand[RuntimeReleaseObjectGroupResult](d.client, "Runtime.releaseObjectGroup", params)
}

func (d RuntimeDomain) RunIfWaitingForDebugger(params ...RuntimeRunIfWaitingForDebuggerParams) (RuntimeRunIfWaitingForDebuggerResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return RuntimeRunIfWaitingForDebuggerResult{}, err
	}
	return sendCDPCommand[RuntimeRunIfWaitingForDebuggerResult](d.client, "Runtime.runIfWaitingForDebugger", p)
}

func (d RuntimeDomain) RunScript(params RuntimeRunScriptParams) (RuntimeRunScriptResult, error) {
	return sendCDPCommand[RuntimeRunScriptResult](d.client, "Runtime.runScript", params)
}

func (d RuntimeDomain) SetAsyncCallStackDepth(params RuntimeSetAsyncCallStackDepthParams) (RuntimeSetAsyncCallStackDepthResult, error) {
	return sendCDPCommand[RuntimeSetAsyncCallStackDepthResult](d.client, "Runtime.setAsyncCallStackDepth", params)
}

func (d RuntimeDomain) SetCustomObjectFormatterEnabled(params RuntimeSetCustomObjectFormatterEnabledParams) (RuntimeSetCustomObjectFormatterEnabledResult, error) {
	return sendCDPCommand[RuntimeSetCustomObjectFormatterEnabledResult](d.client, "Runtime.setCustomObjectFormatterEnabled", params)
}

func (d RuntimeDomain) SetMaxCallStackSizeToCapture(params RuntimeSetMaxCallStackSizeToCaptureParams) (RuntimeSetMaxCallStackSizeToCaptureResult, error) {
	return sendCDPCommand[RuntimeSetMaxCallStackSizeToCaptureResult](d.client, "Runtime.setMaxCallStackSizeToCapture", params)
}

func (d RuntimeDomain) TerminateExecution(params ...RuntimeTerminateExecutionParams) (RuntimeTerminateExecutionResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return RuntimeTerminateExecutionResult{}, err
	}
	return sendCDPCommand[RuntimeTerminateExecutionResult](d.client, "Runtime.terminateExecution", p)
}

func (d RuntimeDomain) AddBinding(params RuntimeAddBindingParams) (RuntimeAddBindingResult, error) {
	return sendCDPCommand[RuntimeAddBindingResult](d.client, "Runtime.addBinding", params)
}

func (d RuntimeDomain) RemoveBinding(params RuntimeRemoveBindingParams) (RuntimeRemoveBindingResult, error) {
	return sendCDPCommand[RuntimeRemoveBindingResult](d.client, "Runtime.removeBinding", params)
}

func (d RuntimeDomain) GetExceptionDetails(params RuntimeGetExceptionDetailsParams) (RuntimeGetExceptionDetailsResult, error) {
	return sendCDPCommand[RuntimeGetExceptionDetailsResult](d.client, "Runtime.getExceptionDetails", params)
}

type SchemaDomain struct{ client *ModCDPClient }

func (d SchemaDomain) GetDomains(params ...SchemaGetDomainsParams) (SchemaGetDomainsResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return SchemaGetDomainsResult{}, err
	}
	return sendCDPCommand[SchemaGetDomainsResult](d.client, "Schema.getDomains", p)
}

type SecurityDomain struct {
	client *ModCDPClient
	On     SecurityEvents
}

func (d SecurityDomain) Disable(params ...SecurityDisableParams) (SecurityDisableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return SecurityDisableResult{}, err
	}
	return sendCDPCommand[SecurityDisableResult](d.client, "Security.disable", p)
}

func (d SecurityDomain) Enable(params ...SecurityEnableParams) (SecurityEnableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return SecurityEnableResult{}, err
	}
	return sendCDPCommand[SecurityEnableResult](d.client, "Security.enable", p)
}

func (d SecurityDomain) SetIgnoreCertificateErrors(params SecuritySetIgnoreCertificateErrorsParams) (SecuritySetIgnoreCertificateErrorsResult, error) {
	return sendCDPCommand[SecuritySetIgnoreCertificateErrorsResult](d.client, "Security.setIgnoreCertificateErrors", params)
}

func (d SecurityDomain) HandleCertificateError(params SecurityHandleCertificateErrorParams) (SecurityHandleCertificateErrorResult, error) {
	return sendCDPCommand[SecurityHandleCertificateErrorResult](d.client, "Security.handleCertificateError", params)
}

func (d SecurityDomain) SetOverrideCertificateErrors(params SecuritySetOverrideCertificateErrorsParams) (SecuritySetOverrideCertificateErrorsResult, error) {
	return sendCDPCommand[SecuritySetOverrideCertificateErrorsResult](d.client, "Security.setOverrideCertificateErrors", params)
}

type ServiceWorkerDomain struct {
	client *ModCDPClient
	On     ServiceWorkerEvents
}

func (d ServiceWorkerDomain) DeliverPushMessage(params ServiceWorkerDeliverPushMessageParams) (ServiceWorkerDeliverPushMessageResult, error) {
	return sendCDPCommand[ServiceWorkerDeliverPushMessageResult](d.client, "ServiceWorker.deliverPushMessage", params)
}

func (d ServiceWorkerDomain) Disable(params ...ServiceWorkerDisableParams) (ServiceWorkerDisableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return ServiceWorkerDisableResult{}, err
	}
	return sendCDPCommand[ServiceWorkerDisableResult](d.client, "ServiceWorker.disable", p)
}

func (d ServiceWorkerDomain) DispatchSyncEvent(params ServiceWorkerDispatchSyncEventParams) (ServiceWorkerDispatchSyncEventResult, error) {
	return sendCDPCommand[ServiceWorkerDispatchSyncEventResult](d.client, "ServiceWorker.dispatchSyncEvent", params)
}

func (d ServiceWorkerDomain) DispatchPeriodicSyncEvent(params ServiceWorkerDispatchPeriodicSyncEventParams) (ServiceWorkerDispatchPeriodicSyncEventResult, error) {
	return sendCDPCommand[ServiceWorkerDispatchPeriodicSyncEventResult](d.client, "ServiceWorker.dispatchPeriodicSyncEvent", params)
}

func (d ServiceWorkerDomain) Enable(params ...ServiceWorkerEnableParams) (ServiceWorkerEnableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return ServiceWorkerEnableResult{}, err
	}
	return sendCDPCommand[ServiceWorkerEnableResult](d.client, "ServiceWorker.enable", p)
}

func (d ServiceWorkerDomain) SetForceUpdateOnPageLoad(params ServiceWorkerSetForceUpdateOnPageLoadParams) (ServiceWorkerSetForceUpdateOnPageLoadResult, error) {
	return sendCDPCommand[ServiceWorkerSetForceUpdateOnPageLoadResult](d.client, "ServiceWorker.setForceUpdateOnPageLoad", params)
}

func (d ServiceWorkerDomain) SkipWaiting(params ServiceWorkerSkipWaitingParams) (ServiceWorkerSkipWaitingResult, error) {
	return sendCDPCommand[ServiceWorkerSkipWaitingResult](d.client, "ServiceWorker.skipWaiting", params)
}

func (d ServiceWorkerDomain) StartWorker(params ServiceWorkerStartWorkerParams) (ServiceWorkerStartWorkerResult, error) {
	return sendCDPCommand[ServiceWorkerStartWorkerResult](d.client, "ServiceWorker.startWorker", params)
}

func (d ServiceWorkerDomain) StopAllWorkers(params ...ServiceWorkerStopAllWorkersParams) (ServiceWorkerStopAllWorkersResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return ServiceWorkerStopAllWorkersResult{}, err
	}
	return sendCDPCommand[ServiceWorkerStopAllWorkersResult](d.client, "ServiceWorker.stopAllWorkers", p)
}

func (d ServiceWorkerDomain) StopWorker(params ServiceWorkerStopWorkerParams) (ServiceWorkerStopWorkerResult, error) {
	return sendCDPCommand[ServiceWorkerStopWorkerResult](d.client, "ServiceWorker.stopWorker", params)
}

func (d ServiceWorkerDomain) Unregister(params ServiceWorkerUnregisterParams) (ServiceWorkerUnregisterResult, error) {
	return sendCDPCommand[ServiceWorkerUnregisterResult](d.client, "ServiceWorker.unregister", params)
}

func (d ServiceWorkerDomain) UpdateRegistration(params ServiceWorkerUpdateRegistrationParams) (ServiceWorkerUpdateRegistrationResult, error) {
	return sendCDPCommand[ServiceWorkerUpdateRegistrationResult](d.client, "ServiceWorker.updateRegistration", params)
}

type SmartCardEmulationDomain struct {
	client *ModCDPClient
	On     SmartCardEmulationEvents
}

func (d SmartCardEmulationDomain) Enable(params ...SmartCardEmulationEnableParams) (SmartCardEmulationEnableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return SmartCardEmulationEnableResult{}, err
	}
	return sendCDPCommand[SmartCardEmulationEnableResult](d.client, "SmartCardEmulation.enable", p)
}

func (d SmartCardEmulationDomain) Disable(params ...SmartCardEmulationDisableParams) (SmartCardEmulationDisableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return SmartCardEmulationDisableResult{}, err
	}
	return sendCDPCommand[SmartCardEmulationDisableResult](d.client, "SmartCardEmulation.disable", p)
}

func (d SmartCardEmulationDomain) ReportEstablishContextResult(params SmartCardEmulationReportEstablishContextResultParams) (SmartCardEmulationReportEstablishContextResultResult, error) {
	return sendCDPCommand[SmartCardEmulationReportEstablishContextResultResult](d.client, "SmartCardEmulation.reportEstablishContextResult", params)
}

func (d SmartCardEmulationDomain) ReportReleaseContextResult(params SmartCardEmulationReportReleaseContextResultParams) (SmartCardEmulationReportReleaseContextResultResult, error) {
	return sendCDPCommand[SmartCardEmulationReportReleaseContextResultResult](d.client, "SmartCardEmulation.reportReleaseContextResult", params)
}

func (d SmartCardEmulationDomain) ReportListReadersResult(params SmartCardEmulationReportListReadersResultParams) (SmartCardEmulationReportListReadersResultResult, error) {
	return sendCDPCommand[SmartCardEmulationReportListReadersResultResult](d.client, "SmartCardEmulation.reportListReadersResult", params)
}

func (d SmartCardEmulationDomain) ReportGetStatusChangeResult(params SmartCardEmulationReportGetStatusChangeResultParams) (SmartCardEmulationReportGetStatusChangeResultResult, error) {
	return sendCDPCommand[SmartCardEmulationReportGetStatusChangeResultResult](d.client, "SmartCardEmulation.reportGetStatusChangeResult", params)
}

func (d SmartCardEmulationDomain) ReportBeginTransactionResult(params SmartCardEmulationReportBeginTransactionResultParams) (SmartCardEmulationReportBeginTransactionResultResult, error) {
	return sendCDPCommand[SmartCardEmulationReportBeginTransactionResultResult](d.client, "SmartCardEmulation.reportBeginTransactionResult", params)
}

func (d SmartCardEmulationDomain) ReportPlainResult(params SmartCardEmulationReportPlainResultParams) (SmartCardEmulationReportPlainResultResult, error) {
	return sendCDPCommand[SmartCardEmulationReportPlainResultResult](d.client, "SmartCardEmulation.reportPlainResult", params)
}

func (d SmartCardEmulationDomain) ReportConnectResult(params SmartCardEmulationReportConnectResultParams) (SmartCardEmulationReportConnectResultResult, error) {
	return sendCDPCommand[SmartCardEmulationReportConnectResultResult](d.client, "SmartCardEmulation.reportConnectResult", params)
}

func (d SmartCardEmulationDomain) ReportDataResult(params SmartCardEmulationReportDataResultParams) (SmartCardEmulationReportDataResultResult, error) {
	return sendCDPCommand[SmartCardEmulationReportDataResultResult](d.client, "SmartCardEmulation.reportDataResult", params)
}

func (d SmartCardEmulationDomain) ReportStatusResult(params SmartCardEmulationReportStatusResultParams) (SmartCardEmulationReportStatusResultResult, error) {
	return sendCDPCommand[SmartCardEmulationReportStatusResultResult](d.client, "SmartCardEmulation.reportStatusResult", params)
}

func (d SmartCardEmulationDomain) ReportError(params SmartCardEmulationReportErrorParams) (SmartCardEmulationReportErrorResult, error) {
	return sendCDPCommand[SmartCardEmulationReportErrorResult](d.client, "SmartCardEmulation.reportError", params)
}

type StorageDomain struct {
	client *ModCDPClient
	On     StorageEvents
}

func (d StorageDomain) GetStorageKeyForFrame(params StorageGetStorageKeyForFrameParams) (StorageGetStorageKeyForFrameResult, error) {
	return sendCDPCommand[StorageGetStorageKeyForFrameResult](d.client, "Storage.getStorageKeyForFrame", params)
}

func (d StorageDomain) GetStorageKey(params ...StorageGetStorageKeyParams) (StorageGetStorageKeyResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return StorageGetStorageKeyResult{}, err
	}
	return sendCDPCommand[StorageGetStorageKeyResult](d.client, "Storage.getStorageKey", p)
}

func (d StorageDomain) ClearDataForOrigin(params StorageClearDataForOriginParams) (StorageClearDataForOriginResult, error) {
	return sendCDPCommand[StorageClearDataForOriginResult](d.client, "Storage.clearDataForOrigin", params)
}

func (d StorageDomain) ClearDataForStorageKey(params StorageClearDataForStorageKeyParams) (StorageClearDataForStorageKeyResult, error) {
	return sendCDPCommand[StorageClearDataForStorageKeyResult](d.client, "Storage.clearDataForStorageKey", params)
}

func (d StorageDomain) GetCookies(params ...StorageGetCookiesParams) (StorageGetCookiesResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return StorageGetCookiesResult{}, err
	}
	return sendCDPCommand[StorageGetCookiesResult](d.client, "Storage.getCookies", p)
}

func (d StorageDomain) SetCookies(params StorageSetCookiesParams) (StorageSetCookiesResult, error) {
	return sendCDPCommand[StorageSetCookiesResult](d.client, "Storage.setCookies", params)
}

func (d StorageDomain) ClearCookies(params ...StorageClearCookiesParams) (StorageClearCookiesResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return StorageClearCookiesResult{}, err
	}
	return sendCDPCommand[StorageClearCookiesResult](d.client, "Storage.clearCookies", p)
}

func (d StorageDomain) GetUsageAndQuota(params StorageGetUsageAndQuotaParams) (StorageGetUsageAndQuotaResult, error) {
	return sendCDPCommand[StorageGetUsageAndQuotaResult](d.client, "Storage.getUsageAndQuota", params)
}

func (d StorageDomain) OverrideQuotaForOrigin(params StorageOverrideQuotaForOriginParams) (StorageOverrideQuotaForOriginResult, error) {
	return sendCDPCommand[StorageOverrideQuotaForOriginResult](d.client, "Storage.overrideQuotaForOrigin", params)
}

func (d StorageDomain) TrackCacheStorageForOrigin(params StorageTrackCacheStorageForOriginParams) (StorageTrackCacheStorageForOriginResult, error) {
	return sendCDPCommand[StorageTrackCacheStorageForOriginResult](d.client, "Storage.trackCacheStorageForOrigin", params)
}

func (d StorageDomain) TrackCacheStorageForStorageKey(params StorageTrackCacheStorageForStorageKeyParams) (StorageTrackCacheStorageForStorageKeyResult, error) {
	return sendCDPCommand[StorageTrackCacheStorageForStorageKeyResult](d.client, "Storage.trackCacheStorageForStorageKey", params)
}

func (d StorageDomain) TrackIndexedDBForOrigin(params StorageTrackIndexedDBForOriginParams) (StorageTrackIndexedDBForOriginResult, error) {
	return sendCDPCommand[StorageTrackIndexedDBForOriginResult](d.client, "Storage.trackIndexedDBForOrigin", params)
}

func (d StorageDomain) TrackIndexedDBForStorageKey(params StorageTrackIndexedDBForStorageKeyParams) (StorageTrackIndexedDBForStorageKeyResult, error) {
	return sendCDPCommand[StorageTrackIndexedDBForStorageKeyResult](d.client, "Storage.trackIndexedDBForStorageKey", params)
}

func (d StorageDomain) UntrackCacheStorageForOrigin(params StorageUntrackCacheStorageForOriginParams) (StorageUntrackCacheStorageForOriginResult, error) {
	return sendCDPCommand[StorageUntrackCacheStorageForOriginResult](d.client, "Storage.untrackCacheStorageForOrigin", params)
}

func (d StorageDomain) UntrackCacheStorageForStorageKey(params StorageUntrackCacheStorageForStorageKeyParams) (StorageUntrackCacheStorageForStorageKeyResult, error) {
	return sendCDPCommand[StorageUntrackCacheStorageForStorageKeyResult](d.client, "Storage.untrackCacheStorageForStorageKey", params)
}

func (d StorageDomain) UntrackIndexedDBForOrigin(params StorageUntrackIndexedDBForOriginParams) (StorageUntrackIndexedDBForOriginResult, error) {
	return sendCDPCommand[StorageUntrackIndexedDBForOriginResult](d.client, "Storage.untrackIndexedDBForOrigin", params)
}

func (d StorageDomain) UntrackIndexedDBForStorageKey(params StorageUntrackIndexedDBForStorageKeyParams) (StorageUntrackIndexedDBForStorageKeyResult, error) {
	return sendCDPCommand[StorageUntrackIndexedDBForStorageKeyResult](d.client, "Storage.untrackIndexedDBForStorageKey", params)
}

func (d StorageDomain) GetTrustTokens(params ...StorageGetTrustTokensParams) (StorageGetTrustTokensResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return StorageGetTrustTokensResult{}, err
	}
	return sendCDPCommand[StorageGetTrustTokensResult](d.client, "Storage.getTrustTokens", p)
}

func (d StorageDomain) ClearTrustTokens(params StorageClearTrustTokensParams) (StorageClearTrustTokensResult, error) {
	return sendCDPCommand[StorageClearTrustTokensResult](d.client, "Storage.clearTrustTokens", params)
}

func (d StorageDomain) GetInterestGroupDetails(params StorageGetInterestGroupDetailsParams) (StorageGetInterestGroupDetailsResult, error) {
	return sendCDPCommand[StorageGetInterestGroupDetailsResult](d.client, "Storage.getInterestGroupDetails", params)
}

func (d StorageDomain) SetInterestGroupTracking(params StorageSetInterestGroupTrackingParams) (StorageSetInterestGroupTrackingResult, error) {
	return sendCDPCommand[StorageSetInterestGroupTrackingResult](d.client, "Storage.setInterestGroupTracking", params)
}

func (d StorageDomain) SetInterestGroupAuctionTracking(params StorageSetInterestGroupAuctionTrackingParams) (StorageSetInterestGroupAuctionTrackingResult, error) {
	return sendCDPCommand[StorageSetInterestGroupAuctionTrackingResult](d.client, "Storage.setInterestGroupAuctionTracking", params)
}

func (d StorageDomain) GetSharedStorageMetadata(params StorageGetSharedStorageMetadataParams) (StorageGetSharedStorageMetadataResult, error) {
	return sendCDPCommand[StorageGetSharedStorageMetadataResult](d.client, "Storage.getSharedStorageMetadata", params)
}

func (d StorageDomain) GetSharedStorageEntries(params StorageGetSharedStorageEntriesParams) (StorageGetSharedStorageEntriesResult, error) {
	return sendCDPCommand[StorageGetSharedStorageEntriesResult](d.client, "Storage.getSharedStorageEntries", params)
}

func (d StorageDomain) SetSharedStorageEntry(params StorageSetSharedStorageEntryParams) (StorageSetSharedStorageEntryResult, error) {
	return sendCDPCommand[StorageSetSharedStorageEntryResult](d.client, "Storage.setSharedStorageEntry", params)
}

func (d StorageDomain) DeleteSharedStorageEntry(params StorageDeleteSharedStorageEntryParams) (StorageDeleteSharedStorageEntryResult, error) {
	return sendCDPCommand[StorageDeleteSharedStorageEntryResult](d.client, "Storage.deleteSharedStorageEntry", params)
}

func (d StorageDomain) ClearSharedStorageEntries(params StorageClearSharedStorageEntriesParams) (StorageClearSharedStorageEntriesResult, error) {
	return sendCDPCommand[StorageClearSharedStorageEntriesResult](d.client, "Storage.clearSharedStorageEntries", params)
}

func (d StorageDomain) ResetSharedStorageBudget(params StorageResetSharedStorageBudgetParams) (StorageResetSharedStorageBudgetResult, error) {
	return sendCDPCommand[StorageResetSharedStorageBudgetResult](d.client, "Storage.resetSharedStorageBudget", params)
}

func (d StorageDomain) SetSharedStorageTracking(params StorageSetSharedStorageTrackingParams) (StorageSetSharedStorageTrackingResult, error) {
	return sendCDPCommand[StorageSetSharedStorageTrackingResult](d.client, "Storage.setSharedStorageTracking", params)
}

func (d StorageDomain) SetStorageBucketTracking(params StorageSetStorageBucketTrackingParams) (StorageSetStorageBucketTrackingResult, error) {
	return sendCDPCommand[StorageSetStorageBucketTrackingResult](d.client, "Storage.setStorageBucketTracking", params)
}

func (d StorageDomain) DeleteStorageBucket(params StorageDeleteStorageBucketParams) (StorageDeleteStorageBucketResult, error) {
	return sendCDPCommand[StorageDeleteStorageBucketResult](d.client, "Storage.deleteStorageBucket", params)
}

func (d StorageDomain) RunBounceTrackingMitigations(params ...StorageRunBounceTrackingMitigationsParams) (StorageRunBounceTrackingMitigationsResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return StorageRunBounceTrackingMitigationsResult{}, err
	}
	return sendCDPCommand[StorageRunBounceTrackingMitigationsResult](d.client, "Storage.runBounceTrackingMitigations", p)
}

func (d StorageDomain) SetAttributionReportingLocalTestingMode(params StorageSetAttributionReportingLocalTestingModeParams) (StorageSetAttributionReportingLocalTestingModeResult, error) {
	return sendCDPCommand[StorageSetAttributionReportingLocalTestingModeResult](d.client, "Storage.setAttributionReportingLocalTestingMode", params)
}

func (d StorageDomain) SetAttributionReportingTracking(params StorageSetAttributionReportingTrackingParams) (StorageSetAttributionReportingTrackingResult, error) {
	return sendCDPCommand[StorageSetAttributionReportingTrackingResult](d.client, "Storage.setAttributionReportingTracking", params)
}

func (d StorageDomain) SendPendingAttributionReports(params ...StorageSendPendingAttributionReportsParams) (StorageSendPendingAttributionReportsResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return StorageSendPendingAttributionReportsResult{}, err
	}
	return sendCDPCommand[StorageSendPendingAttributionReportsResult](d.client, "Storage.sendPendingAttributionReports", p)
}

func (d StorageDomain) GetRelatedWebsiteSets(params ...StorageGetRelatedWebsiteSetsParams) (StorageGetRelatedWebsiteSetsResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return StorageGetRelatedWebsiteSetsResult{}, err
	}
	return sendCDPCommand[StorageGetRelatedWebsiteSetsResult](d.client, "Storage.getRelatedWebsiteSets", p)
}

func (d StorageDomain) GetAffectedUrlsForThirdPartyCookieMetadata(params StorageGetAffectedUrlsForThirdPartyCookieMetadataParams) (StorageGetAffectedUrlsForThirdPartyCookieMetadataResult, error) {
	return sendCDPCommand[StorageGetAffectedUrlsForThirdPartyCookieMetadataResult](d.client, "Storage.getAffectedUrlsForThirdPartyCookieMetadata", params)
}

func (d StorageDomain) SetProtectedAudienceKAnonymity(params StorageSetProtectedAudienceKAnonymityParams) (StorageSetProtectedAudienceKAnonymityResult, error) {
	return sendCDPCommand[StorageSetProtectedAudienceKAnonymityResult](d.client, "Storage.setProtectedAudienceKAnonymity", params)
}

type SystemInfoDomain struct{ client *ModCDPClient }

func (d SystemInfoDomain) GetInfo(params ...SystemInfoGetInfoParams) (SystemInfoGetInfoResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return SystemInfoGetInfoResult{}, err
	}
	return sendCDPCommand[SystemInfoGetInfoResult](d.client, "SystemInfo.getInfo", p)
}

func (d SystemInfoDomain) GetFeatureState(params SystemInfoGetFeatureStateParams) (SystemInfoGetFeatureStateResult, error) {
	return sendCDPCommand[SystemInfoGetFeatureStateResult](d.client, "SystemInfo.getFeatureState", params)
}

func (d SystemInfoDomain) GetProcessInfo(params ...SystemInfoGetProcessInfoParams) (SystemInfoGetProcessInfoResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return SystemInfoGetProcessInfoResult{}, err
	}
	return sendCDPCommand[SystemInfoGetProcessInfoResult](d.client, "SystemInfo.getProcessInfo", p)
}

type TargetDomain struct {
	client *ModCDPClient
	On     TargetEvents
}

func (d TargetDomain) ActivateTarget(params TargetActivateTargetParams) (TargetActivateTargetResult, error) {
	return sendCDPCommand[TargetActivateTargetResult](d.client, "Target.activateTarget", params)
}

func (d TargetDomain) AttachToTarget(params TargetAttachToTargetParams) (TargetAttachToTargetResult, error) {
	return sendCDPCommand[TargetAttachToTargetResult](d.client, "Target.attachToTarget", params)
}

func (d TargetDomain) AttachToBrowserTarget(params ...TargetAttachToBrowserTargetParams) (TargetAttachToBrowserTargetResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return TargetAttachToBrowserTargetResult{}, err
	}
	return sendCDPCommand[TargetAttachToBrowserTargetResult](d.client, "Target.attachToBrowserTarget", p)
}

func (d TargetDomain) CloseTarget(params TargetCloseTargetParams) (TargetCloseTargetResult, error) {
	return sendCDPCommand[TargetCloseTargetResult](d.client, "Target.closeTarget", params)
}

func (d TargetDomain) ExposeDevToolsProtocol(params TargetExposeDevToolsProtocolParams) (TargetExposeDevToolsProtocolResult, error) {
	return sendCDPCommand[TargetExposeDevToolsProtocolResult](d.client, "Target.exposeDevToolsProtocol", params)
}

func (d TargetDomain) CreateBrowserContext(params ...TargetCreateBrowserContextParams) (TargetCreateBrowserContextResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return TargetCreateBrowserContextResult{}, err
	}
	return sendCDPCommand[TargetCreateBrowserContextResult](d.client, "Target.createBrowserContext", p)
}

func (d TargetDomain) GetBrowserContexts(params ...TargetGetBrowserContextsParams) (TargetGetBrowserContextsResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return TargetGetBrowserContextsResult{}, err
	}
	return sendCDPCommand[TargetGetBrowserContextsResult](d.client, "Target.getBrowserContexts", p)
}

func (d TargetDomain) CreateTarget(params TargetCreateTargetParams) (TargetCreateTargetResult, error) {
	return sendCDPCommand[TargetCreateTargetResult](d.client, "Target.createTarget", params)
}

func (d TargetDomain) DetachFromTarget(params ...TargetDetachFromTargetParams) (TargetDetachFromTargetResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return TargetDetachFromTargetResult{}, err
	}
	return sendCDPCommand[TargetDetachFromTargetResult](d.client, "Target.detachFromTarget", p)
}

func (d TargetDomain) DisposeBrowserContext(params TargetDisposeBrowserContextParams) (TargetDisposeBrowserContextResult, error) {
	return sendCDPCommand[TargetDisposeBrowserContextResult](d.client, "Target.disposeBrowserContext", params)
}

func (d TargetDomain) GetTargetInfo(params ...TargetGetTargetInfoParams) (TargetGetTargetInfoResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return TargetGetTargetInfoResult{}, err
	}
	return sendCDPCommand[TargetGetTargetInfoResult](d.client, "Target.getTargetInfo", p)
}

func (d TargetDomain) GetTargets(params ...TargetGetTargetsParams) (TargetGetTargetsResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return TargetGetTargetsResult{}, err
	}
	return sendCDPCommand[TargetGetTargetsResult](d.client, "Target.getTargets", p)
}

func (d TargetDomain) SendMessageToTarget(params TargetSendMessageToTargetParams) (TargetSendMessageToTargetResult, error) {
	return sendCDPCommand[TargetSendMessageToTargetResult](d.client, "Target.sendMessageToTarget", params)
}

func (d TargetDomain) SetAutoAttach(params TargetSetAutoAttachParams) (TargetSetAutoAttachResult, error) {
	return sendCDPCommand[TargetSetAutoAttachResult](d.client, "Target.setAutoAttach", params)
}

func (d TargetDomain) AutoAttachRelated(params TargetAutoAttachRelatedParams) (TargetAutoAttachRelatedResult, error) {
	return sendCDPCommand[TargetAutoAttachRelatedResult](d.client, "Target.autoAttachRelated", params)
}

func (d TargetDomain) SetDiscoverTargets(params TargetSetDiscoverTargetsParams) (TargetSetDiscoverTargetsResult, error) {
	return sendCDPCommand[TargetSetDiscoverTargetsResult](d.client, "Target.setDiscoverTargets", params)
}

func (d TargetDomain) SetRemoteLocations(params TargetSetRemoteLocationsParams) (TargetSetRemoteLocationsResult, error) {
	return sendCDPCommand[TargetSetRemoteLocationsResult](d.client, "Target.setRemoteLocations", params)
}

func (d TargetDomain) GetDevToolsTarget(params TargetGetDevToolsTargetParams) (TargetGetDevToolsTargetResult, error) {
	return sendCDPCommand[TargetGetDevToolsTargetResult](d.client, "Target.getDevToolsTarget", params)
}

func (d TargetDomain) OpenDevTools(params TargetOpenDevToolsParams) (TargetOpenDevToolsResult, error) {
	return sendCDPCommand[TargetOpenDevToolsResult](d.client, "Target.openDevTools", params)
}

type TetheringDomain struct {
	client *ModCDPClient
	On     TetheringEvents
}

func (d TetheringDomain) Bind(params TetheringBindParams) (TetheringBindResult, error) {
	return sendCDPCommand[TetheringBindResult](d.client, "Tethering.bind", params)
}

func (d TetheringDomain) Unbind(params TetheringUnbindParams) (TetheringUnbindResult, error) {
	return sendCDPCommand[TetheringUnbindResult](d.client, "Tethering.unbind", params)
}

type TracingDomain struct {
	client *ModCDPClient
	On     TracingEvents
}

func (d TracingDomain) End(params ...TracingEndParams) (TracingEndResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return TracingEndResult{}, err
	}
	return sendCDPCommand[TracingEndResult](d.client, "Tracing.end", p)
}

func (d TracingDomain) GetCategories(params ...TracingGetCategoriesParams) (TracingGetCategoriesResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return TracingGetCategoriesResult{}, err
	}
	return sendCDPCommand[TracingGetCategoriesResult](d.client, "Tracing.getCategories", p)
}

func (d TracingDomain) GetTrackEventDescriptor(params ...TracingGetTrackEventDescriptorParams) (TracingGetTrackEventDescriptorResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return TracingGetTrackEventDescriptorResult{}, err
	}
	return sendCDPCommand[TracingGetTrackEventDescriptorResult](d.client, "Tracing.getTrackEventDescriptor", p)
}

func (d TracingDomain) RecordClockSyncMarker(params TracingRecordClockSyncMarkerParams) (TracingRecordClockSyncMarkerResult, error) {
	return sendCDPCommand[TracingRecordClockSyncMarkerResult](d.client, "Tracing.recordClockSyncMarker", params)
}

func (d TracingDomain) RequestMemoryDump(params ...TracingRequestMemoryDumpParams) (TracingRequestMemoryDumpResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return TracingRequestMemoryDumpResult{}, err
	}
	return sendCDPCommand[TracingRequestMemoryDumpResult](d.client, "Tracing.requestMemoryDump", p)
}

func (d TracingDomain) Start(params ...TracingStartParams) (TracingStartResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return TracingStartResult{}, err
	}
	return sendCDPCommand[TracingStartResult](d.client, "Tracing.start", p)
}

type WebAudioDomain struct {
	client *ModCDPClient
	On     WebAudioEvents
}

func (d WebAudioDomain) Enable(params ...WebAudioEnableParams) (WebAudioEnableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return WebAudioEnableResult{}, err
	}
	return sendCDPCommand[WebAudioEnableResult](d.client, "WebAudio.enable", p)
}

func (d WebAudioDomain) Disable(params ...WebAudioDisableParams) (WebAudioDisableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return WebAudioDisableResult{}, err
	}
	return sendCDPCommand[WebAudioDisableResult](d.client, "WebAudio.disable", p)
}

func (d WebAudioDomain) GetRealtimeData(params WebAudioGetRealtimeDataParams) (WebAudioGetRealtimeDataResult, error) {
	return sendCDPCommand[WebAudioGetRealtimeDataResult](d.client, "WebAudio.getRealtimeData", params)
}

type WebAuthnDomain struct {
	client *ModCDPClient
	On     WebAuthnEvents
}

func (d WebAuthnDomain) Enable(params ...WebAuthnEnableParams) (WebAuthnEnableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return WebAuthnEnableResult{}, err
	}
	return sendCDPCommand[WebAuthnEnableResult](d.client, "WebAuthn.enable", p)
}

func (d WebAuthnDomain) Disable(params ...WebAuthnDisableParams) (WebAuthnDisableResult, error) {
	p, err := optionalCDPParams(params)
	if err != nil {
		return WebAuthnDisableResult{}, err
	}
	return sendCDPCommand[WebAuthnDisableResult](d.client, "WebAuthn.disable", p)
}

func (d WebAuthnDomain) AddVirtualAuthenticator(params WebAuthnAddVirtualAuthenticatorParams) (WebAuthnAddVirtualAuthenticatorResult, error) {
	return sendCDPCommand[WebAuthnAddVirtualAuthenticatorResult](d.client, "WebAuthn.addVirtualAuthenticator", params)
}

func (d WebAuthnDomain) SetResponseOverrideBits(params WebAuthnSetResponseOverrideBitsParams) (WebAuthnSetResponseOverrideBitsResult, error) {
	return sendCDPCommand[WebAuthnSetResponseOverrideBitsResult](d.client, "WebAuthn.setResponseOverrideBits", params)
}

func (d WebAuthnDomain) RemoveVirtualAuthenticator(params WebAuthnRemoveVirtualAuthenticatorParams) (WebAuthnRemoveVirtualAuthenticatorResult, error) {
	return sendCDPCommand[WebAuthnRemoveVirtualAuthenticatorResult](d.client, "WebAuthn.removeVirtualAuthenticator", params)
}

func (d WebAuthnDomain) AddCredential(params WebAuthnAddCredentialParams) (WebAuthnAddCredentialResult, error) {
	return sendCDPCommand[WebAuthnAddCredentialResult](d.client, "WebAuthn.addCredential", params)
}

func (d WebAuthnDomain) GetCredential(params WebAuthnGetCredentialParams) (WebAuthnGetCredentialResult, error) {
	return sendCDPCommand[WebAuthnGetCredentialResult](d.client, "WebAuthn.getCredential", params)
}

func (d WebAuthnDomain) GetCredentials(params WebAuthnGetCredentialsParams) (WebAuthnGetCredentialsResult, error) {
	return sendCDPCommand[WebAuthnGetCredentialsResult](d.client, "WebAuthn.getCredentials", params)
}

func (d WebAuthnDomain) RemoveCredential(params WebAuthnRemoveCredentialParams) (WebAuthnRemoveCredentialResult, error) {
	return sendCDPCommand[WebAuthnRemoveCredentialResult](d.client, "WebAuthn.removeCredential", params)
}

func (d WebAuthnDomain) ClearCredentials(params WebAuthnClearCredentialsParams) (WebAuthnClearCredentialsResult, error) {
	return sendCDPCommand[WebAuthnClearCredentialsResult](d.client, "WebAuthn.clearCredentials", params)
}

func (d WebAuthnDomain) SetUserVerified(params WebAuthnSetUserVerifiedParams) (WebAuthnSetUserVerifiedResult, error) {
	return sendCDPCommand[WebAuthnSetUserVerifiedResult](d.client, "WebAuthn.setUserVerified", params)
}

func (d WebAuthnDomain) SetAutomaticPresenceSimulation(params WebAuthnSetAutomaticPresenceSimulationParams) (WebAuthnSetAutomaticPresenceSimulationResult, error) {
	return sendCDPCommand[WebAuthnSetAutomaticPresenceSimulationResult](d.client, "WebAuthn.setAutomaticPresenceSimulation", params)
}

func (d WebAuthnDomain) SetCredentialProperties(params WebAuthnSetCredentialPropertiesParams) (WebAuthnSetCredentialPropertiesResult, error) {
	return sendCDPCommand[WebAuthnSetCredentialPropertiesResult](d.client, "WebAuthn.setCredentialProperties", params)
}

type AccessibilityEvents struct{ client *ModCDPClient }

func (e AccessibilityEvents) LoadComplete(handler func(AccessibilityLoadCompleteEvent)) {
	onCDPEvent(e.client, "Accessibility.loadComplete", handler)
}

func (e AccessibilityEvents) NodesUpdated(handler func(AccessibilityNodesUpdatedEvent)) {
	onCDPEvent(e.client, "Accessibility.nodesUpdated", handler)
}

type AnimationEvents struct{ client *ModCDPClient }

func (e AnimationEvents) AnimationCanceled(handler func(AnimationAnimationCanceledEvent)) {
	onCDPEvent(e.client, "Animation.animationCanceled", handler)
}

func (e AnimationEvents) AnimationCreated(handler func(AnimationAnimationCreatedEvent)) {
	onCDPEvent(e.client, "Animation.animationCreated", handler)
}

func (e AnimationEvents) AnimationStarted(handler func(AnimationAnimationStartedEvent)) {
	onCDPEvent(e.client, "Animation.animationStarted", handler)
}

func (e AnimationEvents) AnimationUpdated(handler func(AnimationAnimationUpdatedEvent)) {
	onCDPEvent(e.client, "Animation.animationUpdated", handler)
}

type AuditsEvents struct{ client *ModCDPClient }

func (e AuditsEvents) IssueAdded(handler func(AuditsIssueAddedEvent)) {
	onCDPEvent(e.client, "Audits.issueAdded", handler)
}

type AutofillEvents struct{ client *ModCDPClient }

func (e AutofillEvents) AddressFormFilled(handler func(AutofillAddressFormFilledEvent)) {
	onCDPEvent(e.client, "Autofill.addressFormFilled", handler)
}

type BackgroundServiceEvents struct{ client *ModCDPClient }

func (e BackgroundServiceEvents) RecordingStateChanged(handler func(BackgroundServiceRecordingStateChangedEvent)) {
	onCDPEvent(e.client, "BackgroundService.recordingStateChanged", handler)
}

func (e BackgroundServiceEvents) BackgroundServiceEventReceived(handler func(BackgroundServiceBackgroundServiceEventReceivedEvent)) {
	onCDPEvent(e.client, "BackgroundService.backgroundServiceEventReceived", handler)
}

type BluetoothEmulationEvents struct{ client *ModCDPClient }

func (e BluetoothEmulationEvents) GattOperationReceived(handler func(BluetoothEmulationGattOperationReceivedEvent)) {
	onCDPEvent(e.client, "BluetoothEmulation.gattOperationReceived", handler)
}

func (e BluetoothEmulationEvents) CharacteristicOperationReceived(handler func(BluetoothEmulationCharacteristicOperationReceivedEvent)) {
	onCDPEvent(e.client, "BluetoothEmulation.characteristicOperationReceived", handler)
}

func (e BluetoothEmulationEvents) DescriptorOperationReceived(handler func(BluetoothEmulationDescriptorOperationReceivedEvent)) {
	onCDPEvent(e.client, "BluetoothEmulation.descriptorOperationReceived", handler)
}

type BrowserEvents struct{ client *ModCDPClient }

func (e BrowserEvents) DownloadWillBegin(handler func(BrowserDownloadWillBeginEvent)) {
	onCDPEvent(e.client, "Browser.downloadWillBegin", handler)
}

func (e BrowserEvents) DownloadProgress(handler func(BrowserDownloadProgressEvent)) {
	onCDPEvent(e.client, "Browser.downloadProgress", handler)
}

type CSSEvents struct{ client *ModCDPClient }

func (e CSSEvents) FontsUpdated(handler func(CSSFontsUpdatedEvent)) {
	onCDPEvent(e.client, "CSS.fontsUpdated", handler)
}

func (e CSSEvents) MediaQueryResultChanged(handler func(CSSMediaQueryResultChangedEvent)) {
	onCDPEvent(e.client, "CSS.mediaQueryResultChanged", handler)
}

func (e CSSEvents) StyleSheetAdded(handler func(CSSStyleSheetAddedEvent)) {
	onCDPEvent(e.client, "CSS.styleSheetAdded", handler)
}

func (e CSSEvents) StyleSheetChanged(handler func(CSSStyleSheetChangedEvent)) {
	onCDPEvent(e.client, "CSS.styleSheetChanged", handler)
}

func (e CSSEvents) StyleSheetRemoved(handler func(CSSStyleSheetRemovedEvent)) {
	onCDPEvent(e.client, "CSS.styleSheetRemoved", handler)
}

func (e CSSEvents) ComputedStyleUpdated(handler func(CSSComputedStyleUpdatedEvent)) {
	onCDPEvent(e.client, "CSS.computedStyleUpdated", handler)
}

type CastEvents struct{ client *ModCDPClient }

func (e CastEvents) SinksUpdated(handler func(CastSinksUpdatedEvent)) {
	onCDPEvent(e.client, "Cast.sinksUpdated", handler)
}

func (e CastEvents) IssueUpdated(handler func(CastIssueUpdatedEvent)) {
	onCDPEvent(e.client, "Cast.issueUpdated", handler)
}

type ConsoleEvents struct{ client *ModCDPClient }

func (e ConsoleEvents) MessageAdded(handler func(ConsoleMessageAddedEvent)) {
	onCDPEvent(e.client, "Console.messageAdded", handler)
}

type DOMEvents struct{ client *ModCDPClient }

func (e DOMEvents) AttributeModified(handler func(DOMAttributeModifiedEvent)) {
	onCDPEvent(e.client, "DOM.attributeModified", handler)
}

func (e DOMEvents) AdoptedStyleSheetsModified(handler func(DOMAdoptedStyleSheetsModifiedEvent)) {
	onCDPEvent(e.client, "DOM.adoptedStyleSheetsModified", handler)
}

func (e DOMEvents) AttributeRemoved(handler func(DOMAttributeRemovedEvent)) {
	onCDPEvent(e.client, "DOM.attributeRemoved", handler)
}

func (e DOMEvents) CharacterDataModified(handler func(DOMCharacterDataModifiedEvent)) {
	onCDPEvent(e.client, "DOM.characterDataModified", handler)
}

func (e DOMEvents) ChildNodeCountUpdated(handler func(DOMChildNodeCountUpdatedEvent)) {
	onCDPEvent(e.client, "DOM.childNodeCountUpdated", handler)
}

func (e DOMEvents) ChildNodeInserted(handler func(DOMChildNodeInsertedEvent)) {
	onCDPEvent(e.client, "DOM.childNodeInserted", handler)
}

func (e DOMEvents) ChildNodeRemoved(handler func(DOMChildNodeRemovedEvent)) {
	onCDPEvent(e.client, "DOM.childNodeRemoved", handler)
}

func (e DOMEvents) DistributedNodesUpdated(handler func(DOMDistributedNodesUpdatedEvent)) {
	onCDPEvent(e.client, "DOM.distributedNodesUpdated", handler)
}

func (e DOMEvents) DocumentUpdated(handler func(DOMDocumentUpdatedEvent)) {
	onCDPEvent(e.client, "DOM.documentUpdated", handler)
}

func (e DOMEvents) InlineStyleInvalidated(handler func(DOMInlineStyleInvalidatedEvent)) {
	onCDPEvent(e.client, "DOM.inlineStyleInvalidated", handler)
}

func (e DOMEvents) PseudoElementAdded(handler func(DOMPseudoElementAddedEvent)) {
	onCDPEvent(e.client, "DOM.pseudoElementAdded", handler)
}

func (e DOMEvents) TopLayerElementsUpdated(handler func(DOMTopLayerElementsUpdatedEvent)) {
	onCDPEvent(e.client, "DOM.topLayerElementsUpdated", handler)
}

func (e DOMEvents) ScrollableFlagUpdated(handler func(DOMScrollableFlagUpdatedEvent)) {
	onCDPEvent(e.client, "DOM.scrollableFlagUpdated", handler)
}

func (e DOMEvents) AdRelatedStateUpdated(handler func(DOMAdRelatedStateUpdatedEvent)) {
	onCDPEvent(e.client, "DOM.adRelatedStateUpdated", handler)
}

func (e DOMEvents) AffectedByStartingStylesFlagUpdated(handler func(DOMAffectedByStartingStylesFlagUpdatedEvent)) {
	onCDPEvent(e.client, "DOM.affectedByStartingStylesFlagUpdated", handler)
}

func (e DOMEvents) PseudoElementRemoved(handler func(DOMPseudoElementRemovedEvent)) {
	onCDPEvent(e.client, "DOM.pseudoElementRemoved", handler)
}

func (e DOMEvents) SetChildNodes(handler func(DOMSetChildNodesEvent)) {
	onCDPEvent(e.client, "DOM.setChildNodes", handler)
}

func (e DOMEvents) ShadowRootPopped(handler func(DOMShadowRootPoppedEvent)) {
	onCDPEvent(e.client, "DOM.shadowRootPopped", handler)
}

func (e DOMEvents) ShadowRootPushed(handler func(DOMShadowRootPushedEvent)) {
	onCDPEvent(e.client, "DOM.shadowRootPushed", handler)
}

type DOMStorageEvents struct{ client *ModCDPClient }

func (e DOMStorageEvents) DOMStorageItemAdded(handler func(DOMStorageDOMStorageItemAddedEvent)) {
	onCDPEvent(e.client, "DOMStorage.domStorageItemAdded", handler)
}

func (e DOMStorageEvents) DOMStorageItemRemoved(handler func(DOMStorageDOMStorageItemRemovedEvent)) {
	onCDPEvent(e.client, "DOMStorage.domStorageItemRemoved", handler)
}

func (e DOMStorageEvents) DOMStorageItemUpdated(handler func(DOMStorageDOMStorageItemUpdatedEvent)) {
	onCDPEvent(e.client, "DOMStorage.domStorageItemUpdated", handler)
}

func (e DOMStorageEvents) DOMStorageItemsCleared(handler func(DOMStorageDOMStorageItemsClearedEvent)) {
	onCDPEvent(e.client, "DOMStorage.domStorageItemsCleared", handler)
}

type DebuggerEvents struct{ client *ModCDPClient }

func (e DebuggerEvents) BreakpointResolved(handler func(DebuggerBreakpointResolvedEvent)) {
	onCDPEvent(e.client, "Debugger.breakpointResolved", handler)
}

func (e DebuggerEvents) Paused(handler func(DebuggerPausedEvent)) {
	onCDPEvent(e.client, "Debugger.paused", handler)
}

func (e DebuggerEvents) Resumed(handler func(DebuggerResumedEvent)) {
	onCDPEvent(e.client, "Debugger.resumed", handler)
}

func (e DebuggerEvents) ScriptFailedToParse(handler func(DebuggerScriptFailedToParseEvent)) {
	onCDPEvent(e.client, "Debugger.scriptFailedToParse", handler)
}

func (e DebuggerEvents) ScriptParsed(handler func(DebuggerScriptParsedEvent)) {
	onCDPEvent(e.client, "Debugger.scriptParsed", handler)
}

type DeviceAccessEvents struct{ client *ModCDPClient }

func (e DeviceAccessEvents) DeviceRequestPrompted(handler func(DeviceAccessDeviceRequestPromptedEvent)) {
	onCDPEvent(e.client, "DeviceAccess.deviceRequestPrompted", handler)
}

type EmulationEvents struct{ client *ModCDPClient }

func (e EmulationEvents) VirtualTimeBudgetExpired(handler func(EmulationVirtualTimeBudgetExpiredEvent)) {
	onCDPEvent(e.client, "Emulation.virtualTimeBudgetExpired", handler)
}

func (e EmulationEvents) ScreenOrientationLockChanged(handler func(EmulationScreenOrientationLockChangedEvent)) {
	onCDPEvent(e.client, "Emulation.screenOrientationLockChanged", handler)
}

type FedCmEvents struct{ client *ModCDPClient }

func (e FedCmEvents) DialogShown(handler func(FedCmDialogShownEvent)) {
	onCDPEvent(e.client, "FedCm.dialogShown", handler)
}

func (e FedCmEvents) DialogClosed(handler func(FedCmDialogClosedEvent)) {
	onCDPEvent(e.client, "FedCm.dialogClosed", handler)
}

type FetchEvents struct{ client *ModCDPClient }

func (e FetchEvents) RequestPaused(handler func(FetchRequestPausedEvent)) {
	onCDPEvent(e.client, "Fetch.requestPaused", handler)
}

func (e FetchEvents) AuthRequired(handler func(FetchAuthRequiredEvent)) {
	onCDPEvent(e.client, "Fetch.authRequired", handler)
}

type HeapProfilerEvents struct{ client *ModCDPClient }

func (e HeapProfilerEvents) AddHeapSnapshotChunk(handler func(HeapProfilerAddHeapSnapshotChunkEvent)) {
	onCDPEvent(e.client, "HeapProfiler.addHeapSnapshotChunk", handler)
}

func (e HeapProfilerEvents) HeapStatsUpdate(handler func(HeapProfilerHeapStatsUpdateEvent)) {
	onCDPEvent(e.client, "HeapProfiler.heapStatsUpdate", handler)
}

func (e HeapProfilerEvents) LastSeenObjectID(handler func(HeapProfilerLastSeenObjectIDEvent)) {
	onCDPEvent(e.client, "HeapProfiler.lastSeenObjectId", handler)
}

func (e HeapProfilerEvents) ReportHeapSnapshotProgress(handler func(HeapProfilerReportHeapSnapshotProgressEvent)) {
	onCDPEvent(e.client, "HeapProfiler.reportHeapSnapshotProgress", handler)
}

func (e HeapProfilerEvents) ResetProfiles(handler func(HeapProfilerResetProfilesEvent)) {
	onCDPEvent(e.client, "HeapProfiler.resetProfiles", handler)
}

type InputEvents struct{ client *ModCDPClient }

func (e InputEvents) DragIntercepted(handler func(InputDragInterceptedEvent)) {
	onCDPEvent(e.client, "Input.dragIntercepted", handler)
}

type InspectorEvents struct{ client *ModCDPClient }

func (e InspectorEvents) Detached(handler func(InspectorDetachedEvent)) {
	onCDPEvent(e.client, "Inspector.detached", handler)
}

func (e InspectorEvents) TargetCrashed(handler func(InspectorTargetCrashedEvent)) {
	onCDPEvent(e.client, "Inspector.targetCrashed", handler)
}

func (e InspectorEvents) TargetReloadedAfterCrash(handler func(InspectorTargetReloadedAfterCrashEvent)) {
	onCDPEvent(e.client, "Inspector.targetReloadedAfterCrash", handler)
}

func (e InspectorEvents) WorkerScriptLoaded(handler func(InspectorWorkerScriptLoadedEvent)) {
	onCDPEvent(e.client, "Inspector.workerScriptLoaded", handler)
}

type LayerTreeEvents struct{ client *ModCDPClient }

func (e LayerTreeEvents) LayerPainted(handler func(LayerTreeLayerPaintedEvent)) {
	onCDPEvent(e.client, "LayerTree.layerPainted", handler)
}

func (e LayerTreeEvents) LayerTreeDidChange(handler func(LayerTreeLayerTreeDidChangeEvent)) {
	onCDPEvent(e.client, "LayerTree.layerTreeDidChange", handler)
}

type LogEvents struct{ client *ModCDPClient }

func (e LogEvents) EntryAdded(handler func(LogEntryAddedEvent)) {
	onCDPEvent(e.client, "Log.entryAdded", handler)
}

type MediaEvents struct{ client *ModCDPClient }

func (e MediaEvents) PlayerPropertiesChanged(handler func(MediaPlayerPropertiesChangedEvent)) {
	onCDPEvent(e.client, "Media.playerPropertiesChanged", handler)
}

func (e MediaEvents) PlayerEventsAdded(handler func(MediaPlayerEventsAddedEvent)) {
	onCDPEvent(e.client, "Media.playerEventsAdded", handler)
}

func (e MediaEvents) PlayerMessagesLogged(handler func(MediaPlayerMessagesLoggedEvent)) {
	onCDPEvent(e.client, "Media.playerMessagesLogged", handler)
}

func (e MediaEvents) PlayerErrorsRaised(handler func(MediaPlayerErrorsRaisedEvent)) {
	onCDPEvent(e.client, "Media.playerErrorsRaised", handler)
}

func (e MediaEvents) PlayerCreated(handler func(MediaPlayerCreatedEvent)) {
	onCDPEvent(e.client, "Media.playerCreated", handler)
}

type NetworkEvents struct{ client *ModCDPClient }

func (e NetworkEvents) DataReceived(handler func(NetworkDataReceivedEvent)) {
	onCDPEvent(e.client, "Network.dataReceived", handler)
}

func (e NetworkEvents) EventSourceMessageReceived(handler func(NetworkEventSourceMessageReceivedEvent)) {
	onCDPEvent(e.client, "Network.eventSourceMessageReceived", handler)
}

func (e NetworkEvents) LoadingFailed(handler func(NetworkLoadingFailedEvent)) {
	onCDPEvent(e.client, "Network.loadingFailed", handler)
}

func (e NetworkEvents) LoadingFinished(handler func(NetworkLoadingFinishedEvent)) {
	onCDPEvent(e.client, "Network.loadingFinished", handler)
}

func (e NetworkEvents) RequestIntercepted(handler func(NetworkRequestInterceptedEvent)) {
	onCDPEvent(e.client, "Network.requestIntercepted", handler)
}

func (e NetworkEvents) RequestServedFromCache(handler func(NetworkRequestServedFromCacheEvent)) {
	onCDPEvent(e.client, "Network.requestServedFromCache", handler)
}

func (e NetworkEvents) RequestWillBeSent(handler func(NetworkRequestWillBeSentEvent)) {
	onCDPEvent(e.client, "Network.requestWillBeSent", handler)
}

func (e NetworkEvents) ResourceChangedPriority(handler func(NetworkResourceChangedPriorityEvent)) {
	onCDPEvent(e.client, "Network.resourceChangedPriority", handler)
}

func (e NetworkEvents) SignedExchangeReceived(handler func(NetworkSignedExchangeReceivedEvent)) {
	onCDPEvent(e.client, "Network.signedExchangeReceived", handler)
}

func (e NetworkEvents) ResponseReceived(handler func(NetworkResponseReceivedEvent)) {
	onCDPEvent(e.client, "Network.responseReceived", handler)
}

func (e NetworkEvents) WebSocketClosed(handler func(NetworkWebSocketClosedEvent)) {
	onCDPEvent(e.client, "Network.webSocketClosed", handler)
}

func (e NetworkEvents) WebSocketCreated(handler func(NetworkWebSocketCreatedEvent)) {
	onCDPEvent(e.client, "Network.webSocketCreated", handler)
}

func (e NetworkEvents) WebSocketFrameError(handler func(NetworkWebSocketFrameErrorEvent)) {
	onCDPEvent(e.client, "Network.webSocketFrameError", handler)
}

func (e NetworkEvents) WebSocketFrameReceived(handler func(NetworkWebSocketFrameReceivedEvent)) {
	onCDPEvent(e.client, "Network.webSocketFrameReceived", handler)
}

func (e NetworkEvents) WebSocketFrameSent(handler func(NetworkWebSocketFrameSentEvent)) {
	onCDPEvent(e.client, "Network.webSocketFrameSent", handler)
}

func (e NetworkEvents) WebSocketHandshakeResponseReceived(handler func(NetworkWebSocketHandshakeResponseReceivedEvent)) {
	onCDPEvent(e.client, "Network.webSocketHandshakeResponseReceived", handler)
}

func (e NetworkEvents) WebSocketWillSendHandshakeRequest(handler func(NetworkWebSocketWillSendHandshakeRequestEvent)) {
	onCDPEvent(e.client, "Network.webSocketWillSendHandshakeRequest", handler)
}

func (e NetworkEvents) WebTransportCreated(handler func(NetworkWebTransportCreatedEvent)) {
	onCDPEvent(e.client, "Network.webTransportCreated", handler)
}

func (e NetworkEvents) WebTransportConnectionEstablished(handler func(NetworkWebTransportConnectionEstablishedEvent)) {
	onCDPEvent(e.client, "Network.webTransportConnectionEstablished", handler)
}

func (e NetworkEvents) WebTransportClosed(handler func(NetworkWebTransportClosedEvent)) {
	onCDPEvent(e.client, "Network.webTransportClosed", handler)
}

func (e NetworkEvents) DirectTCPSocketCreated(handler func(NetworkDirectTCPSocketCreatedEvent)) {
	onCDPEvent(e.client, "Network.directTCPSocketCreated", handler)
}

func (e NetworkEvents) DirectTCPSocketOpened(handler func(NetworkDirectTCPSocketOpenedEvent)) {
	onCDPEvent(e.client, "Network.directTCPSocketOpened", handler)
}

func (e NetworkEvents) DirectTCPSocketAborted(handler func(NetworkDirectTCPSocketAbortedEvent)) {
	onCDPEvent(e.client, "Network.directTCPSocketAborted", handler)
}

func (e NetworkEvents) DirectTCPSocketClosed(handler func(NetworkDirectTCPSocketClosedEvent)) {
	onCDPEvent(e.client, "Network.directTCPSocketClosed", handler)
}

func (e NetworkEvents) DirectTCPSocketChunkSent(handler func(NetworkDirectTCPSocketChunkSentEvent)) {
	onCDPEvent(e.client, "Network.directTCPSocketChunkSent", handler)
}

func (e NetworkEvents) DirectTCPSocketChunkReceived(handler func(NetworkDirectTCPSocketChunkReceivedEvent)) {
	onCDPEvent(e.client, "Network.directTCPSocketChunkReceived", handler)
}

func (e NetworkEvents) DirectUDPSocketJoinedMulticastGroup(handler func(NetworkDirectUDPSocketJoinedMulticastGroupEvent)) {
	onCDPEvent(e.client, "Network.directUDPSocketJoinedMulticastGroup", handler)
}

func (e NetworkEvents) DirectUDPSocketLeftMulticastGroup(handler func(NetworkDirectUDPSocketLeftMulticastGroupEvent)) {
	onCDPEvent(e.client, "Network.directUDPSocketLeftMulticastGroup", handler)
}

func (e NetworkEvents) DirectUDPSocketCreated(handler func(NetworkDirectUDPSocketCreatedEvent)) {
	onCDPEvent(e.client, "Network.directUDPSocketCreated", handler)
}

func (e NetworkEvents) DirectUDPSocketOpened(handler func(NetworkDirectUDPSocketOpenedEvent)) {
	onCDPEvent(e.client, "Network.directUDPSocketOpened", handler)
}

func (e NetworkEvents) DirectUDPSocketAborted(handler func(NetworkDirectUDPSocketAbortedEvent)) {
	onCDPEvent(e.client, "Network.directUDPSocketAborted", handler)
}

func (e NetworkEvents) DirectUDPSocketClosed(handler func(NetworkDirectUDPSocketClosedEvent)) {
	onCDPEvent(e.client, "Network.directUDPSocketClosed", handler)
}

func (e NetworkEvents) DirectUDPSocketChunkSent(handler func(NetworkDirectUDPSocketChunkSentEvent)) {
	onCDPEvent(e.client, "Network.directUDPSocketChunkSent", handler)
}

func (e NetworkEvents) DirectUDPSocketChunkReceived(handler func(NetworkDirectUDPSocketChunkReceivedEvent)) {
	onCDPEvent(e.client, "Network.directUDPSocketChunkReceived", handler)
}

func (e NetworkEvents) RequestWillBeSentExtraInfo(handler func(NetworkRequestWillBeSentExtraInfoEvent)) {
	onCDPEvent(e.client, "Network.requestWillBeSentExtraInfo", handler)
}

func (e NetworkEvents) ResponseReceivedExtraInfo(handler func(NetworkResponseReceivedExtraInfoEvent)) {
	onCDPEvent(e.client, "Network.responseReceivedExtraInfo", handler)
}

func (e NetworkEvents) ResponseReceivedEarlyHints(handler func(NetworkResponseReceivedEarlyHintsEvent)) {
	onCDPEvent(e.client, "Network.responseReceivedEarlyHints", handler)
}

func (e NetworkEvents) TrustTokenOperationDone(handler func(NetworkTrustTokenOperationDoneEvent)) {
	onCDPEvent(e.client, "Network.trustTokenOperationDone", handler)
}

func (e NetworkEvents) PolicyUpdated(handler func(NetworkPolicyUpdatedEvent)) {
	onCDPEvent(e.client, "Network.policyUpdated", handler)
}

func (e NetworkEvents) ReportingAPIReportAdded(handler func(NetworkReportingAPIReportAddedEvent)) {
	onCDPEvent(e.client, "Network.reportingApiReportAdded", handler)
}

func (e NetworkEvents) ReportingAPIReportUpdated(handler func(NetworkReportingAPIReportUpdatedEvent)) {
	onCDPEvent(e.client, "Network.reportingApiReportUpdated", handler)
}

func (e NetworkEvents) ReportingAPIEndpointsChangedForOrigin(handler func(NetworkReportingAPIEndpointsChangedForOriginEvent)) {
	onCDPEvent(e.client, "Network.reportingApiEndpointsChangedForOrigin", handler)
}

func (e NetworkEvents) DeviceBoundSessionsAdded(handler func(NetworkDeviceBoundSessionsAddedEvent)) {
	onCDPEvent(e.client, "Network.deviceBoundSessionsAdded", handler)
}

func (e NetworkEvents) DeviceBoundSessionEventOccurred(handler func(NetworkDeviceBoundSessionEventOccurredEvent)) {
	onCDPEvent(e.client, "Network.deviceBoundSessionEventOccurred", handler)
}

type OverlayEvents struct{ client *ModCDPClient }

func (e OverlayEvents) InspectNodeRequested(handler func(OverlayInspectNodeRequestedEvent)) {
	onCDPEvent(e.client, "Overlay.inspectNodeRequested", handler)
}

func (e OverlayEvents) NodeHighlightRequested(handler func(OverlayNodeHighlightRequestedEvent)) {
	onCDPEvent(e.client, "Overlay.nodeHighlightRequested", handler)
}

func (e OverlayEvents) ScreenshotRequested(handler func(OverlayScreenshotRequestedEvent)) {
	onCDPEvent(e.client, "Overlay.screenshotRequested", handler)
}

func (e OverlayEvents) InspectPanelShowRequested(handler func(OverlayInspectPanelShowRequestedEvent)) {
	onCDPEvent(e.client, "Overlay.inspectPanelShowRequested", handler)
}

func (e OverlayEvents) InspectedElementWindowRestored(handler func(OverlayInspectedElementWindowRestoredEvent)) {
	onCDPEvent(e.client, "Overlay.inspectedElementWindowRestored", handler)
}

func (e OverlayEvents) InspectModeCanceled(handler func(OverlayInspectModeCanceledEvent)) {
	onCDPEvent(e.client, "Overlay.inspectModeCanceled", handler)
}

type PageEvents struct{ client *ModCDPClient }

func (e PageEvents) DOMContentEventFired(handler func(PageDOMContentEventFiredEvent)) {
	onCDPEvent(e.client, "Page.domContentEventFired", handler)
}

func (e PageEvents) FileChooserOpened(handler func(PageFileChooserOpenedEvent)) {
	onCDPEvent(e.client, "Page.fileChooserOpened", handler)
}

func (e PageEvents) FrameAttached(handler func(PageFrameAttachedEvent)) {
	onCDPEvent(e.client, "Page.frameAttached", handler)
}

func (e PageEvents) FrameClearedScheduledNavigation(handler func(PageFrameClearedScheduledNavigationEvent)) {
	onCDPEvent(e.client, "Page.frameClearedScheduledNavigation", handler)
}

func (e PageEvents) FrameDetached(handler func(PageFrameDetachedEvent)) {
	onCDPEvent(e.client, "Page.frameDetached", handler)
}

func (e PageEvents) FrameSubtreeWillBeDetached(handler func(PageFrameSubtreeWillBeDetachedEvent)) {
	onCDPEvent(e.client, "Page.frameSubtreeWillBeDetached", handler)
}

func (e PageEvents) FrameNavigated(handler func(PageFrameNavigatedEvent)) {
	onCDPEvent(e.client, "Page.frameNavigated", handler)
}

func (e PageEvents) DocumentOpened(handler func(PageDocumentOpenedEvent)) {
	onCDPEvent(e.client, "Page.documentOpened", handler)
}

func (e PageEvents) FrameResized(handler func(PageFrameResizedEvent)) {
	onCDPEvent(e.client, "Page.frameResized", handler)
}

func (e PageEvents) FrameStartedNavigating(handler func(PageFrameStartedNavigatingEvent)) {
	onCDPEvent(e.client, "Page.frameStartedNavigating", handler)
}

func (e PageEvents) FrameRequestedNavigation(handler func(PageFrameRequestedNavigationEvent)) {
	onCDPEvent(e.client, "Page.frameRequestedNavigation", handler)
}

func (e PageEvents) FrameScheduledNavigation(handler func(PageFrameScheduledNavigationEvent)) {
	onCDPEvent(e.client, "Page.frameScheduledNavigation", handler)
}

func (e PageEvents) FrameStartedLoading(handler func(PageFrameStartedLoadingEvent)) {
	onCDPEvent(e.client, "Page.frameStartedLoading", handler)
}

func (e PageEvents) FrameStoppedLoading(handler func(PageFrameStoppedLoadingEvent)) {
	onCDPEvent(e.client, "Page.frameStoppedLoading", handler)
}

func (e PageEvents) DownloadWillBegin(handler func(PageDownloadWillBeginEvent)) {
	onCDPEvent(e.client, "Page.downloadWillBegin", handler)
}

func (e PageEvents) DownloadProgress(handler func(PageDownloadProgressEvent)) {
	onCDPEvent(e.client, "Page.downloadProgress", handler)
}

func (e PageEvents) InterstitialHidden(handler func(PageInterstitialHiddenEvent)) {
	onCDPEvent(e.client, "Page.interstitialHidden", handler)
}

func (e PageEvents) InterstitialShown(handler func(PageInterstitialShownEvent)) {
	onCDPEvent(e.client, "Page.interstitialShown", handler)
}

func (e PageEvents) JavascriptDialogClosed(handler func(PageJavascriptDialogClosedEvent)) {
	onCDPEvent(e.client, "Page.javascriptDialogClosed", handler)
}

func (e PageEvents) JavascriptDialogOpening(handler func(PageJavascriptDialogOpeningEvent)) {
	onCDPEvent(e.client, "Page.javascriptDialogOpening", handler)
}

func (e PageEvents) LifecycleEvent(handler func(PageLifecycleEventEvent)) {
	onCDPEvent(e.client, "Page.lifecycleEvent", handler)
}

func (e PageEvents) BackForwardCacheNotUsed(handler func(PageBackForwardCacheNotUsedEvent)) {
	onCDPEvent(e.client, "Page.backForwardCacheNotUsed", handler)
}

func (e PageEvents) LoadEventFired(handler func(PageLoadEventFiredEvent)) {
	onCDPEvent(e.client, "Page.loadEventFired", handler)
}

func (e PageEvents) NavigatedWithinDocument(handler func(PageNavigatedWithinDocumentEvent)) {
	onCDPEvent(e.client, "Page.navigatedWithinDocument", handler)
}

func (e PageEvents) ScreencastFrame(handler func(PageScreencastFrameEvent)) {
	onCDPEvent(e.client, "Page.screencastFrame", handler)
}

func (e PageEvents) ScreencastVisibilityChanged(handler func(PageScreencastVisibilityChangedEvent)) {
	onCDPEvent(e.client, "Page.screencastVisibilityChanged", handler)
}

func (e PageEvents) WindowOpen(handler func(PageWindowOpenEvent)) {
	onCDPEvent(e.client, "Page.windowOpen", handler)
}

func (e PageEvents) CompilationCacheProduced(handler func(PageCompilationCacheProducedEvent)) {
	onCDPEvent(e.client, "Page.compilationCacheProduced", handler)
}

type PerformanceEvents struct{ client *ModCDPClient }

func (e PerformanceEvents) Metrics(handler func(PerformanceMetricsEvent)) {
	onCDPEvent(e.client, "Performance.metrics", handler)
}

type PerformanceTimelineEvents struct{ client *ModCDPClient }

func (e PerformanceTimelineEvents) TimelineEventAdded(handler func(PerformanceTimelineTimelineEventAddedEvent)) {
	onCDPEvent(e.client, "PerformanceTimeline.timelineEventAdded", handler)
}

type PreloadEvents struct{ client *ModCDPClient }

func (e PreloadEvents) RuleSetUpdated(handler func(PreloadRuleSetUpdatedEvent)) {
	onCDPEvent(e.client, "Preload.ruleSetUpdated", handler)
}

func (e PreloadEvents) RuleSetRemoved(handler func(PreloadRuleSetRemovedEvent)) {
	onCDPEvent(e.client, "Preload.ruleSetRemoved", handler)
}

func (e PreloadEvents) PreloadEnabledStateUpdated(handler func(PreloadPreloadEnabledStateUpdatedEvent)) {
	onCDPEvent(e.client, "Preload.preloadEnabledStateUpdated", handler)
}

func (e PreloadEvents) PrefetchStatusUpdated(handler func(PreloadPrefetchStatusUpdatedEvent)) {
	onCDPEvent(e.client, "Preload.prefetchStatusUpdated", handler)
}

func (e PreloadEvents) PrerenderStatusUpdated(handler func(PreloadPrerenderStatusUpdatedEvent)) {
	onCDPEvent(e.client, "Preload.prerenderStatusUpdated", handler)
}

func (e PreloadEvents) PreloadingAttemptSourcesUpdated(handler func(PreloadPreloadingAttemptSourcesUpdatedEvent)) {
	onCDPEvent(e.client, "Preload.preloadingAttemptSourcesUpdated", handler)
}

type ProfilerEvents struct{ client *ModCDPClient }

func (e ProfilerEvents) ConsoleProfileFinished(handler func(ProfilerConsoleProfileFinishedEvent)) {
	onCDPEvent(e.client, "Profiler.consoleProfileFinished", handler)
}

func (e ProfilerEvents) ConsoleProfileStarted(handler func(ProfilerConsoleProfileStartedEvent)) {
	onCDPEvent(e.client, "Profiler.consoleProfileStarted", handler)
}

func (e ProfilerEvents) PreciseCoverageDeltaUpdate(handler func(ProfilerPreciseCoverageDeltaUpdateEvent)) {
	onCDPEvent(e.client, "Profiler.preciseCoverageDeltaUpdate", handler)
}

type RuntimeEvents struct{ client *ModCDPClient }

func (e RuntimeEvents) BindingCalled(handler func(RuntimeBindingCalledEvent)) {
	onCDPEvent(e.client, "Runtime.bindingCalled", handler)
}

func (e RuntimeEvents) ConsoleAPICalled(handler func(RuntimeConsoleAPICalledEvent)) {
	onCDPEvent(e.client, "Runtime.consoleAPICalled", handler)
}

func (e RuntimeEvents) ExceptionRevoked(handler func(RuntimeExceptionRevokedEvent)) {
	onCDPEvent(e.client, "Runtime.exceptionRevoked", handler)
}

func (e RuntimeEvents) ExceptionThrown(handler func(RuntimeExceptionThrownEvent)) {
	onCDPEvent(e.client, "Runtime.exceptionThrown", handler)
}

func (e RuntimeEvents) ExecutionContextCreated(handler func(RuntimeExecutionContextCreatedEvent)) {
	onCDPEvent(e.client, "Runtime.executionContextCreated", handler)
}

func (e RuntimeEvents) ExecutionContextDestroyed(handler func(RuntimeExecutionContextDestroyedEvent)) {
	onCDPEvent(e.client, "Runtime.executionContextDestroyed", handler)
}

func (e RuntimeEvents) ExecutionContextsCleared(handler func(RuntimeExecutionContextsClearedEvent)) {
	onCDPEvent(e.client, "Runtime.executionContextsCleared", handler)
}

func (e RuntimeEvents) InspectRequested(handler func(RuntimeInspectRequestedEvent)) {
	onCDPEvent(e.client, "Runtime.inspectRequested", handler)
}

type SecurityEvents struct{ client *ModCDPClient }

func (e SecurityEvents) CertificateError(handler func(SecurityCertificateErrorEvent)) {
	onCDPEvent(e.client, "Security.certificateError", handler)
}

func (e SecurityEvents) VisibleSecurityStateChanged(handler func(SecurityVisibleSecurityStateChangedEvent)) {
	onCDPEvent(e.client, "Security.visibleSecurityStateChanged", handler)
}

func (e SecurityEvents) SecurityStateChanged(handler func(SecuritySecurityStateChangedEvent)) {
	onCDPEvent(e.client, "Security.securityStateChanged", handler)
}

type ServiceWorkerEvents struct{ client *ModCDPClient }

func (e ServiceWorkerEvents) WorkerErrorReported(handler func(ServiceWorkerWorkerErrorReportedEvent)) {
	onCDPEvent(e.client, "ServiceWorker.workerErrorReported", handler)
}

func (e ServiceWorkerEvents) WorkerRegistrationUpdated(handler func(ServiceWorkerWorkerRegistrationUpdatedEvent)) {
	onCDPEvent(e.client, "ServiceWorker.workerRegistrationUpdated", handler)
}

func (e ServiceWorkerEvents) WorkerVersionUpdated(handler func(ServiceWorkerWorkerVersionUpdatedEvent)) {
	onCDPEvent(e.client, "ServiceWorker.workerVersionUpdated", handler)
}

type SmartCardEmulationEvents struct{ client *ModCDPClient }

func (e SmartCardEmulationEvents) EstablishContextRequested(handler func(SmartCardEmulationEstablishContextRequestedEvent)) {
	onCDPEvent(e.client, "SmartCardEmulation.establishContextRequested", handler)
}

func (e SmartCardEmulationEvents) ReleaseContextRequested(handler func(SmartCardEmulationReleaseContextRequestedEvent)) {
	onCDPEvent(e.client, "SmartCardEmulation.releaseContextRequested", handler)
}

func (e SmartCardEmulationEvents) ListReadersRequested(handler func(SmartCardEmulationListReadersRequestedEvent)) {
	onCDPEvent(e.client, "SmartCardEmulation.listReadersRequested", handler)
}

func (e SmartCardEmulationEvents) GetStatusChangeRequested(handler func(SmartCardEmulationGetStatusChangeRequestedEvent)) {
	onCDPEvent(e.client, "SmartCardEmulation.getStatusChangeRequested", handler)
}

func (e SmartCardEmulationEvents) CancelRequested(handler func(SmartCardEmulationCancelRequestedEvent)) {
	onCDPEvent(e.client, "SmartCardEmulation.cancelRequested", handler)
}

func (e SmartCardEmulationEvents) ConnectRequested(handler func(SmartCardEmulationConnectRequestedEvent)) {
	onCDPEvent(e.client, "SmartCardEmulation.connectRequested", handler)
}

func (e SmartCardEmulationEvents) DisconnectRequested(handler func(SmartCardEmulationDisconnectRequestedEvent)) {
	onCDPEvent(e.client, "SmartCardEmulation.disconnectRequested", handler)
}

func (e SmartCardEmulationEvents) TransmitRequested(handler func(SmartCardEmulationTransmitRequestedEvent)) {
	onCDPEvent(e.client, "SmartCardEmulation.transmitRequested", handler)
}

func (e SmartCardEmulationEvents) ControlRequested(handler func(SmartCardEmulationControlRequestedEvent)) {
	onCDPEvent(e.client, "SmartCardEmulation.controlRequested", handler)
}

func (e SmartCardEmulationEvents) GetAttribRequested(handler func(SmartCardEmulationGetAttribRequestedEvent)) {
	onCDPEvent(e.client, "SmartCardEmulation.getAttribRequested", handler)
}

func (e SmartCardEmulationEvents) SetAttribRequested(handler func(SmartCardEmulationSetAttribRequestedEvent)) {
	onCDPEvent(e.client, "SmartCardEmulation.setAttribRequested", handler)
}

func (e SmartCardEmulationEvents) StatusRequested(handler func(SmartCardEmulationStatusRequestedEvent)) {
	onCDPEvent(e.client, "SmartCardEmulation.statusRequested", handler)
}

func (e SmartCardEmulationEvents) BeginTransactionRequested(handler func(SmartCardEmulationBeginTransactionRequestedEvent)) {
	onCDPEvent(e.client, "SmartCardEmulation.beginTransactionRequested", handler)
}

func (e SmartCardEmulationEvents) EndTransactionRequested(handler func(SmartCardEmulationEndTransactionRequestedEvent)) {
	onCDPEvent(e.client, "SmartCardEmulation.endTransactionRequested", handler)
}

type StorageEvents struct{ client *ModCDPClient }

func (e StorageEvents) CacheStorageContentUpdated(handler func(StorageCacheStorageContentUpdatedEvent)) {
	onCDPEvent(e.client, "Storage.cacheStorageContentUpdated", handler)
}

func (e StorageEvents) CacheStorageListUpdated(handler func(StorageCacheStorageListUpdatedEvent)) {
	onCDPEvent(e.client, "Storage.cacheStorageListUpdated", handler)
}

func (e StorageEvents) IndexedDBContentUpdated(handler func(StorageIndexedDBContentUpdatedEvent)) {
	onCDPEvent(e.client, "Storage.indexedDBContentUpdated", handler)
}

func (e StorageEvents) IndexedDBListUpdated(handler func(StorageIndexedDBListUpdatedEvent)) {
	onCDPEvent(e.client, "Storage.indexedDBListUpdated", handler)
}

func (e StorageEvents) InterestGroupAccessed(handler func(StorageInterestGroupAccessedEvent)) {
	onCDPEvent(e.client, "Storage.interestGroupAccessed", handler)
}

func (e StorageEvents) InterestGroupAuctionEventOccurred(handler func(StorageInterestGroupAuctionEventOccurredEvent)) {
	onCDPEvent(e.client, "Storage.interestGroupAuctionEventOccurred", handler)
}

func (e StorageEvents) InterestGroupAuctionNetworkRequestCreated(handler func(StorageInterestGroupAuctionNetworkRequestCreatedEvent)) {
	onCDPEvent(e.client, "Storage.interestGroupAuctionNetworkRequestCreated", handler)
}

func (e StorageEvents) SharedStorageAccessed(handler func(StorageSharedStorageAccessedEvent)) {
	onCDPEvent(e.client, "Storage.sharedStorageAccessed", handler)
}

func (e StorageEvents) SharedStorageWorkletOperationExecutionFinished(handler func(StorageSharedStorageWorkletOperationExecutionFinishedEvent)) {
	onCDPEvent(e.client, "Storage.sharedStorageWorkletOperationExecutionFinished", handler)
}

func (e StorageEvents) StorageBucketCreatedOrUpdated(handler func(StorageStorageBucketCreatedOrUpdatedEvent)) {
	onCDPEvent(e.client, "Storage.storageBucketCreatedOrUpdated", handler)
}

func (e StorageEvents) StorageBucketDeleted(handler func(StorageStorageBucketDeletedEvent)) {
	onCDPEvent(e.client, "Storage.storageBucketDeleted", handler)
}

func (e StorageEvents) AttributionReportingSourceRegistered(handler func(StorageAttributionReportingSourceRegisteredEvent)) {
	onCDPEvent(e.client, "Storage.attributionReportingSourceRegistered", handler)
}

func (e StorageEvents) AttributionReportingTriggerRegistered(handler func(StorageAttributionReportingTriggerRegisteredEvent)) {
	onCDPEvent(e.client, "Storage.attributionReportingTriggerRegistered", handler)
}

func (e StorageEvents) AttributionReportingReportSent(handler func(StorageAttributionReportingReportSentEvent)) {
	onCDPEvent(e.client, "Storage.attributionReportingReportSent", handler)
}

func (e StorageEvents) AttributionReportingVerboseDebugReportSent(handler func(StorageAttributionReportingVerboseDebugReportSentEvent)) {
	onCDPEvent(e.client, "Storage.attributionReportingVerboseDebugReportSent", handler)
}

type TargetEvents struct{ client *ModCDPClient }

func (e TargetEvents) AttachedToTarget(handler func(TargetAttachedToTargetEvent)) {
	onCDPEvent(e.client, "Target.attachedToTarget", handler)
}

func (e TargetEvents) DetachedFromTarget(handler func(TargetDetachedFromTargetEvent)) {
	onCDPEvent(e.client, "Target.detachedFromTarget", handler)
}

func (e TargetEvents) ReceivedMessageFromTarget(handler func(TargetReceivedMessageFromTargetEvent)) {
	onCDPEvent(e.client, "Target.receivedMessageFromTarget", handler)
}

func (e TargetEvents) TargetCreated(handler func(TargetTargetCreatedEvent)) {
	onCDPEvent(e.client, "Target.targetCreated", handler)
}

func (e TargetEvents) TargetDestroyed(handler func(TargetTargetDestroyedEvent)) {
	onCDPEvent(e.client, "Target.targetDestroyed", handler)
}

func (e TargetEvents) TargetCrashed(handler func(TargetTargetCrashedEvent)) {
	onCDPEvent(e.client, "Target.targetCrashed", handler)
}

func (e TargetEvents) TargetInfoChanged(handler func(TargetTargetInfoChangedEvent)) {
	onCDPEvent(e.client, "Target.targetInfoChanged", handler)
}

type TetheringEvents struct{ client *ModCDPClient }

func (e TetheringEvents) Accepted(handler func(TetheringAcceptedEvent)) {
	onCDPEvent(e.client, "Tethering.accepted", handler)
}

type TracingEvents struct{ client *ModCDPClient }

func (e TracingEvents) BufferUsage(handler func(TracingBufferUsageEvent)) {
	onCDPEvent(e.client, "Tracing.bufferUsage", handler)
}

func (e TracingEvents) DataCollected(handler func(TracingDataCollectedEvent)) {
	onCDPEvent(e.client, "Tracing.dataCollected", handler)
}

func (e TracingEvents) TracingComplete(handler func(TracingTracingCompleteEvent)) {
	onCDPEvent(e.client, "Tracing.tracingComplete", handler)
}

type WebAudioEvents struct{ client *ModCDPClient }

func (e WebAudioEvents) ContextCreated(handler func(WebAudioContextCreatedEvent)) {
	onCDPEvent(e.client, "WebAudio.contextCreated", handler)
}

func (e WebAudioEvents) ContextWillBeDestroyed(handler func(WebAudioContextWillBeDestroyedEvent)) {
	onCDPEvent(e.client, "WebAudio.contextWillBeDestroyed", handler)
}

func (e WebAudioEvents) ContextChanged(handler func(WebAudioContextChangedEvent)) {
	onCDPEvent(e.client, "WebAudio.contextChanged", handler)
}

func (e WebAudioEvents) AudioListenerCreated(handler func(WebAudioAudioListenerCreatedEvent)) {
	onCDPEvent(e.client, "WebAudio.audioListenerCreated", handler)
}

func (e WebAudioEvents) AudioListenerWillBeDestroyed(handler func(WebAudioAudioListenerWillBeDestroyedEvent)) {
	onCDPEvent(e.client, "WebAudio.audioListenerWillBeDestroyed", handler)
}

func (e WebAudioEvents) AudioNodeCreated(handler func(WebAudioAudioNodeCreatedEvent)) {
	onCDPEvent(e.client, "WebAudio.audioNodeCreated", handler)
}

func (e WebAudioEvents) AudioNodeWillBeDestroyed(handler func(WebAudioAudioNodeWillBeDestroyedEvent)) {
	onCDPEvent(e.client, "WebAudio.audioNodeWillBeDestroyed", handler)
}

func (e WebAudioEvents) AudioParamCreated(handler func(WebAudioAudioParamCreatedEvent)) {
	onCDPEvent(e.client, "WebAudio.audioParamCreated", handler)
}

func (e WebAudioEvents) AudioParamWillBeDestroyed(handler func(WebAudioAudioParamWillBeDestroyedEvent)) {
	onCDPEvent(e.client, "WebAudio.audioParamWillBeDestroyed", handler)
}

func (e WebAudioEvents) NodesConnected(handler func(WebAudioNodesConnectedEvent)) {
	onCDPEvent(e.client, "WebAudio.nodesConnected", handler)
}

func (e WebAudioEvents) NodesDisconnected(handler func(WebAudioNodesDisconnectedEvent)) {
	onCDPEvent(e.client, "WebAudio.nodesDisconnected", handler)
}

func (e WebAudioEvents) NodeParamConnected(handler func(WebAudioNodeParamConnectedEvent)) {
	onCDPEvent(e.client, "WebAudio.nodeParamConnected", handler)
}

func (e WebAudioEvents) NodeParamDisconnected(handler func(WebAudioNodeParamDisconnectedEvent)) {
	onCDPEvent(e.client, "WebAudio.nodeParamDisconnected", handler)
}

type WebAuthnEvents struct{ client *ModCDPClient }

func (e WebAuthnEvents) CredentialAdded(handler func(WebAuthnCredentialAddedEvent)) {
	onCDPEvent(e.client, "WebAuthn.credentialAdded", handler)
}

func (e WebAuthnEvents) CredentialDeleted(handler func(WebAuthnCredentialDeletedEvent)) {
	onCDPEvent(e.client, "WebAuthn.credentialDeleted", handler)
}

func (e WebAuthnEvents) CredentialUpdated(handler func(WebAuthnCredentialUpdatedEvent)) {
	onCDPEvent(e.client, "WebAuthn.credentialUpdated", handler)
}

func (e WebAuthnEvents) CredentialAsserted(handler func(WebAuthnCredentialAssertedEvent)) {
	onCDPEvent(e.client, "WebAuthn.credentialAsserted", handler)
}
