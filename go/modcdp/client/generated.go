// Code generated from Chrome DevTools Protocol JSON. DO NOT EDIT.
package client

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
)

func Bool(v bool) *bool          { return &v }
func Int(v int) *int             { return &v }
func Float64(v float64) *float64 { return &v }
func String(v string) *string    { return &v }

func initCDPSurface(c *ModCDPClient) {
	c.Accessibility = AccessibilityDomain{client: c, On: AccessibilityEvents{client: c}}
	c.Animation = AnimationDomain{client: c, On: AnimationEvents{client: c}}
	c.Audits = AuditsDomain{client: c, On: AuditsEvents{client: c}}
	c.Autofill = AutofillDomain{client: c, On: AutofillEvents{client: c}}
	c.BackgroundService = BackgroundServiceDomain{client: c, On: BackgroundServiceEvents{client: c}}
	c.BluetoothEmulation = BluetoothEmulationDomain{client: c, On: BluetoothEmulationEvents{client: c}}
	c.Browser = BrowserDomain{client: c, On: BrowserEvents{client: c}}
	c.CSS = CSSDomain{client: c, On: CSSEvents{client: c}}
	c.CacheStorage = CacheStorageDomain{client: c}
	c.Cast = CastDomain{client: c, On: CastEvents{client: c}}
	c.Console = ConsoleDomain{client: c, On: ConsoleEvents{client: c}}
	c.DOM = DOMDomain{client: c, On: DOMEvents{client: c}}
	c.DOMDebugger = DOMDebuggerDomain{client: c}
	c.DOMSnapshot = DOMSnapshotDomain{client: c}
	c.DOMStorage = DOMStorageDomain{client: c, On: DOMStorageEvents{client: c}}
	c.Debugger = DebuggerDomain{client: c, On: DebuggerEvents{client: c}}
	c.DeviceAccess = DeviceAccessDomain{client: c, On: DeviceAccessEvents{client: c}}
	c.DeviceOrientation = DeviceOrientationDomain{client: c}
	c.Emulation = EmulationDomain{client: c, On: EmulationEvents{client: c}}
	c.EventBreakpoints = EventBreakpointsDomain{client: c}
	c.Extensions = ExtensionsDomain{client: c}
	c.FedCm = FedCmDomain{client: c, On: FedCmEvents{client: c}}
	c.Fetch = FetchDomain{client: c, On: FetchEvents{client: c}}
	c.FileSystem = FileSystemDomain{client: c}
	c.HeadlessExperimental = HeadlessExperimentalDomain{client: c}
	c.HeapProfiler = HeapProfilerDomain{client: c, On: HeapProfilerEvents{client: c}}
	c.IO = IODomain{client: c}
	c.IndexedDB = IndexedDBDomain{client: c}
	c.Input = InputDomain{client: c, On: InputEvents{client: c}}
	c.Inspector = InspectorDomain{client: c, On: InspectorEvents{client: c}}
	c.LayerTree = LayerTreeDomain{client: c, On: LayerTreeEvents{client: c}}
	c.Log = LogDomain{client: c, On: LogEvents{client: c}}
	c.Media = MediaDomain{client: c, On: MediaEvents{client: c}}
	c.Memory = MemoryDomain{client: c}
	c.Network = NetworkDomain{client: c, On: NetworkEvents{client: c}}
	c.Overlay = OverlayDomain{client: c, On: OverlayEvents{client: c}}
	c.PWA = PWADomain{client: c}
	c.Page = PageDomain{client: c, On: PageEvents{client: c}}
	c.Performance = PerformanceDomain{client: c, On: PerformanceEvents{client: c}}
	c.PerformanceTimeline = PerformanceTimelineDomain{client: c, On: PerformanceTimelineEvents{client: c}}
	c.Preload = PreloadDomain{client: c, On: PreloadEvents{client: c}}
	c.Profiler = ProfilerDomain{client: c, On: ProfilerEvents{client: c}}
	c.Runtime = RuntimeDomain{client: c, On: RuntimeEvents{client: c}}
	c.Schema = SchemaDomain{client: c}
	c.Security = SecurityDomain{client: c, On: SecurityEvents{client: c}}
	c.ServiceWorker = ServiceWorkerDomain{client: c, On: ServiceWorkerEvents{client: c}}
	c.SmartCardEmulation = SmartCardEmulationDomain{client: c, On: SmartCardEmulationEvents{client: c}}
	c.Storage = StorageDomain{client: c, On: StorageEvents{client: c}}
	c.SystemInfo = SystemInfoDomain{client: c}
	c.Target = TargetDomain{client: c, On: TargetEvents{client: c}}
	c.Tethering = TetheringDomain{client: c, On: TetheringEvents{client: c}}
	c.Tracing = TracingDomain{client: c, On: TracingEvents{client: c}}
	c.WebAudio = WebAudioDomain{client: c, On: WebAudioEvents{client: c}}
	c.WebAuthn = WebAuthnDomain{client: c, On: WebAuthnEvents{client: c}}
}

func cdpSessionID(params any) string {
	value := reflect.ValueOf(params)
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return ""
		}
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct {
		return ""
	}
	field := value.FieldByName("SessionID")
	if field.IsValid() && field.Kind() == reflect.String {
		return field.String()
	}
	return ""
}

func cdpParamsMap(params any) (map[string]any, error) {
	if params == nil {
		return map[string]any{}, nil
	}
	body, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	out := map[string]any{}
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func sendCDPCommand[T any](client *ModCDPClient, method string, params any) (T, error) {
	var typed T
	if client == nil {
		return typed, fmt.Errorf("client_hydrate_aliases is false; use Send or SendRaw for %s", method)
	}
	rawParams, err := cdpParamsMap(params)
	if err != nil {
		return typed, err
	}
	sessionID := cdpSessionID(params)
	var result any
	if sessionID != "" {
		result, err = client.sendCommand(method, rawParams, sessionID, true)
	} else {
		result, err = client.Send(method, rawParams)
	}
	if err != nil {
		return typed, err
	}
	body, err := json.Marshal(result)
	if err != nil {
		return typed, err
	}
	if err := json.Unmarshal(body, &typed); err != nil {
		return typed, fmt.Errorf("%s result did not match typed result shape: %w", method, err)
	}
	return typed, nil
}

func optionalCDPParams[T any](params []T) (T, error) {
	var zero T
	if len(params) == 0 {
		return zero, nil
	}
	if len(params) == 1 {
		return params[0], nil
	}
	return zero, fmt.Errorf("expected at most one params object")
}

func onCDPEvent[T any](client *ModCDPClient, event string, handler func(T)) {
	client.On(event, func(data any) {
		var typed T
		body, err := json.Marshal(data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[ModCDPClient] %s event could not be encoded for typed handler: %v\n", event, err)
			return
		}
		if err := json.Unmarshal(body, &typed); err != nil {
			fmt.Fprintf(os.Stderr, "[ModCDPClient] %s event did not match typed event shape: %v\n", event, err)
			return
		}
		handler(typed)
	})
}

type AccessibilityAXNodeID string

type AccessibilityAXValueType string

type AccessibilityAXValueSourceType string

type AccessibilityAXValueNativeSourceType string

type AccessibilityAXValueSource struct {
	// What type of source this is.
	Type AccessibilityAXValueSourceType `json:"type"`
	// The value of this property source.
	Value *AccessibilityAXValue `json:"value,omitempty"`
	// The name of the relevant attribute, if any.
	Attribute *string `json:"attribute,omitempty"`
	// The value of the relevant attribute, if any.
	AttributeValue *AccessibilityAXValue `json:"attributeValue,omitempty"`
	// Whether this source is superseded by a higher priority source.
	Superseded *bool `json:"superseded,omitempty"`
	// The native markup source for this value, e.g. a `<label>` element.
	NativeSource *AccessibilityAXValueNativeSourceType `json:"nativeSource,omitempty"`
	// The value, such as a node or node list, of the native source.
	NativeSourceValue *AccessibilityAXValue `json:"nativeSourceValue,omitempty"`
	// Whether the value for this property is invalid.
	Invalid *bool `json:"invalid,omitempty"`
	// Reason for the value being invalid, if it is.
	InvalidReason *string `json:"invalidReason,omitempty"`
}

type AccessibilityAXRelatedNode struct {
	// The BackendNodeId of the related DOM node.
	BackendDOMNodeID DOMBackendNodeID `json:"backendDOMNodeId"`
	// The IDRef value provided, if any.
	Idref *string `json:"idref,omitempty"`
	// The text alternative of this node in the current context.
	Text *string `json:"text,omitempty"`
}

type AccessibilityAXProperty struct {
	// The name of this property.
	Name AccessibilityAXPropertyName `json:"name"`
	// The value of this property.
	Value AccessibilityAXValue `json:"value"`
}

type AccessibilityAXValue struct {
	// The type of this value.
	Type AccessibilityAXValueType `json:"type"`
	// The computed value of this property.
	Value any `json:"value,omitempty"`
	// One or more related nodes, if applicable.
	RelatedNodes []AccessibilityAXRelatedNode `json:"relatedNodes,omitempty"`
	// The sources which contributed to the computation of this property.
	Sources []AccessibilityAXValueSource `json:"sources,omitempty"`
}

type AccessibilityAXPropertyName string

type AccessibilityAXNode struct {
	// Unique identifier for this node.
	NodeID AccessibilityAXNodeID `json:"nodeId"`
	// Whether this node is ignored for accessibility
	Ignored bool `json:"ignored"`
	// Collection of reasons why this node is hidden.
	IgnoredReasons []AccessibilityAXProperty `json:"ignoredReasons,omitempty"`
	// This `Node`'s role, whether explicit or implicit.
	Role *AccessibilityAXValue `json:"role,omitempty"`
	// This `Node`'s Chrome raw role.
	ChromeRole *AccessibilityAXValue `json:"chromeRole,omitempty"`
	// The accessible name for this `Node`.
	Name *AccessibilityAXValue `json:"name,omitempty"`
	// The accessible description for this `Node`.
	Description *AccessibilityAXValue `json:"description,omitempty"`
	// The value for this `Node`.
	Value *AccessibilityAXValue `json:"value,omitempty"`
	// All other properties
	Properties []AccessibilityAXProperty `json:"properties,omitempty"`
	// ID for this node's parent.
	ParentID *AccessibilityAXNodeID `json:"parentId,omitempty"`
	// IDs for each of this node's child nodes.
	ChildIds []AccessibilityAXNodeID `json:"childIds,omitempty"`
	// The backend ID for the associated DOM node, if any.
	BackendDOMNodeID *DOMBackendNodeID `json:"backendDOMNodeId,omitempty"`
	// The frame ID for the frame associated with this nodes document.
	FrameID *PageFrameID `json:"frameId,omitempty"`
}

type AccessibilityDisableParams struct {
	SessionID string `json:"-"`
}

type AccessibilityDisableResult struct {
}

type AccessibilityEnableParams struct {
	SessionID string `json:"-"`
}

type AccessibilityEnableResult struct {
}

type AccessibilityGetPartialAXTreeParams struct {
	SessionID string `json:"-"`
	// Identifier of the node to get the partial accessibility tree for.
	NodeID *DOMNodeID `json:"nodeId,omitempty"`
	// Identifier of the backend node to get the partial accessibility tree for.
	BackendNodeID *DOMBackendNodeID `json:"backendNodeId,omitempty"`
	// JavaScript object id of the node wrapper to get the partial accessibility tree for.
	ObjectID *RuntimeRemoteObjectID `json:"objectId,omitempty"`
	// Whether to fetch this node's ancestors, siblings and children. Defaults to true.
	FetchRelatives *bool `json:"fetchRelatives,omitempty"`
}

type AccessibilityGetPartialAXTreeResult struct {
	// The `Accessibility.AXNode` for this DOM node, if it exists, plus its ancestors, siblings and
	Nodes []AccessibilityAXNode `json:"nodes"`
}

type AccessibilityGetFullAXTreeParams struct {
	SessionID string `json:"-"`
	// The maximum depth at which descendants of the root node should be retrieved.
	Depth *int `json:"depth,omitempty"`
	// The frame for whose document the AX tree should be retrieved.
	FrameID *PageFrameID `json:"frameId,omitempty"`
}

type AccessibilityGetFullAXTreeResult struct {
	Nodes []AccessibilityAXNode `json:"nodes"`
}

type AccessibilityGetRootAXNodeParams struct {
	SessionID string `json:"-"`
	// The frame in whose document the node resides.
	FrameID *PageFrameID `json:"frameId,omitempty"`
}

type AccessibilityGetRootAXNodeResult struct {
	Node AccessibilityAXNode `json:"node"`
}

type AccessibilityGetAXNodeAndAncestorsParams struct {
	SessionID string `json:"-"`
	// Identifier of the node to get.
	NodeID *DOMNodeID `json:"nodeId,omitempty"`
	// Identifier of the backend node to get.
	BackendNodeID *DOMBackendNodeID `json:"backendNodeId,omitempty"`
	// JavaScript object id of the node wrapper to get.
	ObjectID *RuntimeRemoteObjectID `json:"objectId,omitempty"`
}

type AccessibilityGetAXNodeAndAncestorsResult struct {
	Nodes []AccessibilityAXNode `json:"nodes"`
}

type AccessibilityGetChildAXNodesParams struct {
	SessionID string                `json:"-"`
	ID        AccessibilityAXNodeID `json:"id"`
	// The frame in whose document the node resides.
	FrameID *PageFrameID `json:"frameId,omitempty"`
}

type AccessibilityGetChildAXNodesResult struct {
	Nodes []AccessibilityAXNode `json:"nodes"`
}

type AccessibilityQueryAXTreeParams struct {
	SessionID string `json:"-"`
	// Identifier of the node for the root to query.
	NodeID *DOMNodeID `json:"nodeId,omitempty"`
	// Identifier of the backend node for the root to query.
	BackendNodeID *DOMBackendNodeID `json:"backendNodeId,omitempty"`
	// JavaScript object id of the node wrapper for the root to query.
	ObjectID *RuntimeRemoteObjectID `json:"objectId,omitempty"`
	// Find nodes with this computed name.
	AccessibleName *string `json:"accessibleName,omitempty"`
	// Find nodes with this computed role.
	Role *string `json:"role,omitempty"`
}

type AccessibilityQueryAXTreeResult struct {
	// A list of `Accessibility.AXNode` matching the specified attributes,
	Nodes []AccessibilityAXNode `json:"nodes"`
}

type AccessibilityLoadCompleteEvent struct {
	// New document root node.
	Root AccessibilityAXNode `json:"root"`
}

type AccessibilityNodesUpdatedEvent struct {
	// Updated node data.
	Nodes []AccessibilityAXNode `json:"nodes"`
}

type AnimationAnimation struct {
	// `Animation`'s id.
	ID string `json:"id"`
	// `Animation`'s name.
	Name string `json:"name"`
	// `Animation`'s internal paused state.
	PausedState bool `json:"pausedState"`
	// `Animation`'s play state.
	PlayState string `json:"playState"`
	// `Animation`'s playback rate.
	PlaybackRate float64 `json:"playbackRate"`
	// `Animation`'s start time.
	StartTime float64 `json:"startTime"`
	// `Animation`'s current time.
	CurrentTime float64 `json:"currentTime"`
	// Animation type of `Animation`.
	Type string `json:"type"`
	// `Animation`'s source animation node.
	Source *AnimationAnimationEffect `json:"source,omitempty"`
	// A unique ID for `Animation` representing the sources that triggered this CSS
	CSSID *string `json:"cssId,omitempty"`
	// View or scroll timeline
	ViewOrScrollTimeline *AnimationViewOrScrollTimeline `json:"viewOrScrollTimeline,omitempty"`
}

type AnimationViewOrScrollTimeline struct {
	// Scroll container node
	SourceNodeID *DOMBackendNodeID `json:"sourceNodeId,omitempty"`
	// Represents the starting scroll position of the timeline
	StartOffset *float64 `json:"startOffset,omitempty"`
	// Represents the ending scroll position of the timeline
	EndOffset *float64 `json:"endOffset,omitempty"`
	// The element whose principal box's visibility in the
	SubjectNodeID *DOMBackendNodeID `json:"subjectNodeId,omitempty"`
	// Orientation of the scroll
	Axis DOMScrollOrientation `json:"axis"`
}

type AnimationAnimationEffect struct {
	// `AnimationEffect`'s delay.
	Delay float64 `json:"delay"`
	// `AnimationEffect`'s end delay.
	EndDelay float64 `json:"endDelay"`
	// `AnimationEffect`'s iteration start.
	IterationStart float64 `json:"iterationStart"`
	// `AnimationEffect`'s iterations. Omitted if the value is infinite.
	Iterations *float64 `json:"iterations,omitempty"`
	// `AnimationEffect`'s iteration duration.
	Duration float64 `json:"duration"`
	// `AnimationEffect`'s playback direction.
	Direction string `json:"direction"`
	// `AnimationEffect`'s fill mode.
	Fill string `json:"fill"`
	// `AnimationEffect`'s target node.
	BackendNodeID *DOMBackendNodeID `json:"backendNodeId,omitempty"`
	// `AnimationEffect`'s keyframes.
	KeyframesRule *AnimationKeyframesRule `json:"keyframesRule,omitempty"`
	// `AnimationEffect`'s timing function.
	Easing string `json:"easing"`
}

type AnimationKeyframesRule struct {
	// CSS keyframed animation's name.
	Name *string `json:"name,omitempty"`
	// List of animation keyframes.
	Keyframes []AnimationKeyframeStyle `json:"keyframes"`
}

type AnimationKeyframeStyle struct {
	// Keyframe's time offset.
	Offset string `json:"offset"`
	// `AnimationEffect`'s timing function.
	Easing string `json:"easing"`
}

type AnimationDisableParams struct {
	SessionID string `json:"-"`
}

type AnimationDisableResult struct {
}

type AnimationEnableParams struct {
	SessionID string `json:"-"`
}

type AnimationEnableResult struct {
}

type AnimationGetCurrentTimeParams struct {
	SessionID string `json:"-"`
	// Id of animation.
	ID string `json:"id"`
}

type AnimationGetCurrentTimeResult struct {
	// Current time of the page.
	CurrentTime float64 `json:"currentTime"`
}

type AnimationGetPlaybackRateParams struct {
	SessionID string `json:"-"`
}

type AnimationGetPlaybackRateResult struct {
	// Playback rate for animations on page.
	PlaybackRate float64 `json:"playbackRate"`
}

type AnimationReleaseAnimationsParams struct {
	SessionID string `json:"-"`
	// List of animation ids to seek.
	Animations []string `json:"animations"`
}

type AnimationReleaseAnimationsResult struct {
}

type AnimationResolveAnimationParams struct {
	SessionID string `json:"-"`
	// Animation id.
	AnimationID string `json:"animationId"`
}

type AnimationResolveAnimationResult struct {
	// Corresponding remote object.
	RemoteObject RuntimeRemoteObject `json:"remoteObject"`
}

type AnimationSeekAnimationsParams struct {
	SessionID string `json:"-"`
	// List of animation ids to seek.
	Animations []string `json:"animations"`
	// Set the current time of each animation.
	CurrentTime float64 `json:"currentTime"`
}

type AnimationSeekAnimationsResult struct {
}

type AnimationSetPausedParams struct {
	SessionID string `json:"-"`
	// Animations to set the pause state of.
	Animations []string `json:"animations"`
	// Paused state to set to.
	Paused bool `json:"paused"`
}

type AnimationSetPausedResult struct {
}

type AnimationSetPlaybackRateParams struct {
	SessionID string `json:"-"`
	// Playback rate for animations on page
	PlaybackRate float64 `json:"playbackRate"`
}

type AnimationSetPlaybackRateResult struct {
}

type AnimationSetTimingParams struct {
	SessionID string `json:"-"`
	// Animation id.
	AnimationID string `json:"animationId"`
	// Duration of the animation.
	Duration float64 `json:"duration"`
	// Delay of the animation.
	Delay float64 `json:"delay"`
}

type AnimationSetTimingResult struct {
}

type AnimationAnimationCanceledEvent struct {
	// Id of the animation that was cancelled.
	ID string `json:"id"`
}

type AnimationAnimationCreatedEvent struct {
	// Id of the animation that was created.
	ID string `json:"id"`
}

type AnimationAnimationStartedEvent struct {
	// Animation that was started.
	Animation AnimationAnimation `json:"animation"`
}

type AnimationAnimationUpdatedEvent struct {
	// Animation that was updated.
	Animation AnimationAnimation `json:"animation"`
}

type AuditsAffectedCookie struct {
	// The following three properties uniquely identify a cookie
	Name   string `json:"name"`
	Path   string `json:"path"`
	Domain string `json:"domain"`
}

type AuditsAffectedRequest struct {
	// The unique request id.
	RequestID *NetworkRequestID `json:"requestId,omitempty"`
	URL       string            `json:"url"`
}

type AuditsAffectedFrame struct {
	FrameID PageFrameID `json:"frameId"`
}

type AuditsCookieExclusionReason string

type AuditsCookieWarningReason string

type AuditsCookieOperation string

type AuditsInsightType string

type AuditsCookieIssueInsight struct {
	Type AuditsInsightType `json:"type"`
	// Link to table entry in third-party cookie migration readiness list.
	TableEntryURL *string `json:"tableEntryUrl,omitempty"`
}

type AuditsCookieIssueDetails struct {
	// If AffectedCookie is not set then rawCookieLine contains the raw
	Cookie                 *AuditsAffectedCookie         `json:"cookie,omitempty"`
	RawCookieLine          *string                       `json:"rawCookieLine,omitempty"`
	CookieWarningReasons   []AuditsCookieWarningReason   `json:"cookieWarningReasons"`
	CookieExclusionReasons []AuditsCookieExclusionReason `json:"cookieExclusionReasons"`
	// Optionally identifies the site-for-cookies and the cookie url, which
	Operation      AuditsCookieOperation  `json:"operation"`
	SiteForCookies *string                `json:"siteForCookies,omitempty"`
	CookieURL      *string                `json:"cookieUrl,omitempty"`
	Request        *AuditsAffectedRequest `json:"request,omitempty"`
	// The recommended solution to the issue.
	Insight *AuditsCookieIssueInsight `json:"insight,omitempty"`
}

type AuditsPerformanceIssueType string

type AuditsPerformanceIssueDetails struct {
	PerformanceIssueType AuditsPerformanceIssueType `json:"performanceIssueType"`
	SourceCodeLocation   *AuditsSourceCodeLocation  `json:"sourceCodeLocation,omitempty"`
}

type AuditsMixedContentResolutionStatus string

type AuditsMixedContentResourceType string

type AuditsMixedContentIssueDetails struct {
	// The type of resource causing the mixed content issue (css, js, iframe,
	ResourceType *AuditsMixedContentResourceType `json:"resourceType,omitempty"`
	// The way the mixed content issue is being resolved.
	ResolutionStatus AuditsMixedContentResolutionStatus `json:"resolutionStatus"`
	// The unsafe http url causing the mixed content issue.
	InsecureURL string `json:"insecureURL"`
	// The url responsible for the call to an unsafe url.
	MainResourceURL string `json:"mainResourceURL"`
	// The mixed content request.
	Request *AuditsAffectedRequest `json:"request,omitempty"`
	// Optional because not every mixed content issue is necessarily linked to a frame.
	Frame *AuditsAffectedFrame `json:"frame,omitempty"`
}

type AuditsBlockedByResponseReason string

type AuditsBlockedByResponseIssueDetails struct {
	Request      AuditsAffectedRequest         `json:"request"`
	ParentFrame  *AuditsAffectedFrame          `json:"parentFrame,omitempty"`
	BlockedFrame *AuditsAffectedFrame          `json:"blockedFrame,omitempty"`
	Reason       AuditsBlockedByResponseReason `json:"reason"`
}

type AuditsHeavyAdResolutionStatus string

type AuditsHeavyAdReason string

type AuditsHeavyAdIssueDetails struct {
	// The resolution status, either blocking the content or warning.
	Resolution AuditsHeavyAdResolutionStatus `json:"resolution"`
	// The reason the ad was blocked, total network or cpu or peak cpu.
	Reason AuditsHeavyAdReason `json:"reason"`
	// The frame that was blocked.
	Frame AuditsAffectedFrame `json:"frame"`
}

type AuditsContentSecurityPolicyViolationType string

type AuditsSourceCodeLocation struct {
	ScriptID     *RuntimeScriptID `json:"scriptId,omitempty"`
	URL          string           `json:"url"`
	LineNumber   int              `json:"lineNumber"`
	ColumnNumber int              `json:"columnNumber"`
}

type AuditsContentSecurityPolicyIssueDetails struct {
	// The url not included in allowed sources.
	BlockedURL *string `json:"blockedURL,omitempty"`
	// Specific directive that is violated, causing the CSP issue.
	ViolatedDirective                  string                                   `json:"violatedDirective"`
	IsReportOnly                       bool                                     `json:"isReportOnly"`
	ContentSecurityPolicyViolationType AuditsContentSecurityPolicyViolationType `json:"contentSecurityPolicyViolationType"`
	FrameAncestor                      *AuditsAffectedFrame                     `json:"frameAncestor,omitempty"`
	SourceCodeLocation                 *AuditsSourceCodeLocation                `json:"sourceCodeLocation,omitempty"`
	ViolatingNodeID                    *DOMBackendNodeID                        `json:"violatingNodeId,omitempty"`
}

type AuditsSharedArrayBufferIssueType string

type AuditsSharedArrayBufferIssueDetails struct {
	SourceCodeLocation AuditsSourceCodeLocation         `json:"sourceCodeLocation"`
	IsWarning          bool                             `json:"isWarning"`
	Type               AuditsSharedArrayBufferIssueType `json:"type"`
}

type AuditsCorsIssueDetails struct {
	CorsErrorStatus        NetworkCorsErrorStatus      `json:"corsErrorStatus"`
	IsWarning              bool                        `json:"isWarning"`
	Request                AuditsAffectedRequest       `json:"request"`
	Location               *AuditsSourceCodeLocation   `json:"location,omitempty"`
	InitiatorOrigin        *string                     `json:"initiatorOrigin,omitempty"`
	ResourceIPAddressSpace *NetworkIPAddressSpace      `json:"resourceIPAddressSpace,omitempty"`
	ClientSecurityState    *NetworkClientSecurityState `json:"clientSecurityState,omitempty"`
}

type AuditsAttributionReportingIssueType string

type AuditsSharedDictionaryError string

type AuditsSRIMessageSignatureError string

type AuditsUnencodedDigestError string

type AuditsConnectionAllowlistError string

type AuditsAttributionReportingIssueDetails struct {
	ViolationType    AuditsAttributionReportingIssueType `json:"violationType"`
	Request          *AuditsAffectedRequest              `json:"request,omitempty"`
	ViolatingNodeID  *DOMBackendNodeID                   `json:"violatingNodeId,omitempty"`
	InvalidParameter *string                             `json:"invalidParameter,omitempty"`
}

type AuditsQuirksModeIssueDetails struct {
	// If false, it means the document's mode is "quirks"
	IsLimitedQuirksMode bool             `json:"isLimitedQuirksMode"`
	DocumentNodeID      DOMBackendNodeID `json:"documentNodeId"`
	URL                 string           `json:"url"`
	FrameID             PageFrameID      `json:"frameId"`
	LoaderID            NetworkLoaderID  `json:"loaderId"`
}

type AuditsNavigatorUserAgentIssueDetails struct {
	URL      string                    `json:"url"`
	Location *AuditsSourceCodeLocation `json:"location,omitempty"`
}

type AuditsSharedDictionaryIssueDetails struct {
	SharedDictionaryError AuditsSharedDictionaryError `json:"sharedDictionaryError"`
	Request               AuditsAffectedRequest       `json:"request"`
}

type AuditsSRIMessageSignatureIssueDetails struct {
	Error               AuditsSRIMessageSignatureError `json:"error"`
	SignatureBase       string                         `json:"signatureBase"`
	IntegrityAssertions []string                       `json:"integrityAssertions"`
	Request             AuditsAffectedRequest          `json:"request"`
}

type AuditsUnencodedDigestIssueDetails struct {
	Error   AuditsUnencodedDigestError `json:"error"`
	Request AuditsAffectedRequest      `json:"request"`
}

type AuditsConnectionAllowlistIssueDetails struct {
	Error   AuditsConnectionAllowlistError `json:"error"`
	Request AuditsAffectedRequest          `json:"request"`
}

type AuditsGenericIssueErrorType string

type AuditsGenericIssueDetails struct {
	// Issues with the same errorType are aggregated in the frontend.
	ErrorType              AuditsGenericIssueErrorType `json:"errorType"`
	FrameID                *PageFrameID                `json:"frameId,omitempty"`
	ViolatingNodeID        *DOMBackendNodeID           `json:"violatingNodeId,omitempty"`
	ViolatingNodeAttribute *string                     `json:"violatingNodeAttribute,omitempty"`
	Request                *AuditsAffectedRequest      `json:"request,omitempty"`
}

type AuditsDeprecationIssueDetails struct {
	AffectedFrame      *AuditsAffectedFrame     `json:"affectedFrame,omitempty"`
	SourceCodeLocation AuditsSourceCodeLocation `json:"sourceCodeLocation"`
	// One of the deprecation names from third_party/blink/renderer/core/frame/deprecation/deprecation.json5
	Type string `json:"type"`
}

type AuditsBounceTrackingIssueDetails struct {
	TrackingSites []string `json:"trackingSites"`
}

type AuditsCookieDeprecationMetadataIssueDetails struct {
	AllowedSites     []string              `json:"allowedSites"`
	OptOutPercentage float64               `json:"optOutPercentage"`
	IsOptOutTopLevel bool                  `json:"isOptOutTopLevel"`
	Operation        AuditsCookieOperation `json:"operation"`
}

type AuditsClientHintIssueReason string

type AuditsFederatedAuthRequestIssueDetails struct {
	FederatedAuthRequestIssueReason AuditsFederatedAuthRequestIssueReason `json:"federatedAuthRequestIssueReason"`
}

type AuditsFederatedAuthRequestIssueReason string

type AuditsFederatedAuthUserInfoRequestIssueDetails struct {
	FederatedAuthUserInfoRequestIssueReason AuditsFederatedAuthUserInfoRequestIssueReason `json:"federatedAuthUserInfoRequestIssueReason"`
}

type AuditsFederatedAuthUserInfoRequestIssueReason string

type AuditsClientHintIssueDetails struct {
	SourceCodeLocation    AuditsSourceCodeLocation    `json:"sourceCodeLocation"`
	ClientHintIssueReason AuditsClientHintIssueReason `json:"clientHintIssueReason"`
}

type AuditsFailedRequestInfo struct {
	// The URL that failed to load.
	URL string `json:"url"`
	// The failure message for the failed request.
	FailureMessage string            `json:"failureMessage"`
	RequestID      *NetworkRequestID `json:"requestId,omitempty"`
}

type AuditsPartitioningBlobURLInfo string

type AuditsPartitioningBlobURLIssueDetails struct {
	// The BlobURL that failed to load.
	URL string `json:"url"`
	// Additional information about the Partitioning Blob URL issue.
	PartitioningBlobURLInfo AuditsPartitioningBlobURLInfo `json:"partitioningBlobURLInfo"`
}

type AuditsElementAccessibilityIssueReason string

type AuditsElementAccessibilityIssueDetails struct {
	NodeID                          DOMBackendNodeID                      `json:"nodeId"`
	ElementAccessibilityIssueReason AuditsElementAccessibilityIssueReason `json:"elementAccessibilityIssueReason"`
	HasDisallowedAttributes         bool                                  `json:"hasDisallowedAttributes"`
}

type AuditsStyleSheetLoadingIssueReason string

type AuditsStylesheetLoadingIssueDetails struct {
	// Source code position that referenced the failing stylesheet.
	SourceCodeLocation AuditsSourceCodeLocation `json:"sourceCodeLocation"`
	// Reason why the stylesheet couldn't be loaded.
	StyleSheetLoadingIssueReason AuditsStyleSheetLoadingIssueReason `json:"styleSheetLoadingIssueReason"`
	// Contains additional info when the failure was due to a request.
	FailedRequestInfo *AuditsFailedRequestInfo `json:"failedRequestInfo,omitempty"`
}

type AuditsPropertyRuleIssueReason string

type AuditsPropertyRuleIssueDetails struct {
	// Source code position of the property rule.
	SourceCodeLocation AuditsSourceCodeLocation `json:"sourceCodeLocation"`
	// Reason why the property rule was discarded.
	PropertyRuleIssueReason AuditsPropertyRuleIssueReason `json:"propertyRuleIssueReason"`
	// The value of the property rule property that failed to parse
	PropertyValue *string `json:"propertyValue,omitempty"`
}

type AuditsUserReidentificationIssueType string

type AuditsUserReidentificationIssueDetails struct {
	Type AuditsUserReidentificationIssueType `json:"type"`
	// Applies to BlockedFrameNavigation and BlockedSubresource issue types.
	Request *AuditsAffectedRequest `json:"request,omitempty"`
	// Applies to NoisedCanvasReadback issue type.
	SourceCodeLocation *AuditsSourceCodeLocation `json:"sourceCodeLocation,omitempty"`
}

type AuditsPermissionElementIssueType string

type AuditsPermissionElementIssueDetails struct {
	IssueType AuditsPermissionElementIssueType `json:"issueType"`
	// The value of the type attribute.
	Type *string `json:"type,omitempty"`
	// The node ID of the <permission> element.
	NodeID *DOMBackendNodeID `json:"nodeId,omitempty"`
	// True if the issue is a warning, false if it is an error.
	IsWarning *bool `json:"isWarning,omitempty"`
	// Fields for message construction:
	PermissionName *string `json:"permissionName,omitempty"`
	// Used for messages about occlusion
	OccluderNodeInfo *string `json:"occluderNodeInfo,omitempty"`
	// Used for messages about occluder's parent
	OccluderParentNodeInfo *string `json:"occluderParentNodeInfo,omitempty"`
	// Used for messages about activation disabled reason
	DisableReason *string `json:"disableReason,omitempty"`
}

type AuditsAdScriptIdentifier struct {
	// The script's v8 identifier.
	ScriptID RuntimeScriptID `json:"scriptId"`
	// v8's debugging id for the v8::Context.
	DebuggerID RuntimeUniqueDebuggerID `json:"debuggerId"`
	// The script's url (or generated name based on id if inline script).
	Name string `json:"name"`
}

type AuditsAdAncestry struct {
	// The ad-script in the stack when the offending script was loaded. This is
	AdAncestryChain []AuditsAdScriptIdentifier `json:"adAncestryChain"`
	// The filterlist rule that caused the root (last) script in
	RootScriptFilterlistRule *string `json:"rootScriptFilterlistRule,omitempty"`
}

type AuditsSelectivePermissionsInterventionIssueDetails struct {
	// Which API was intervened on.
	APIName string `json:"apiName"`
	// Why the ad script using the API is considered an ad.
	AdAncestry AuditsAdAncestry `json:"adAncestry"`
	// The stack trace at the time of the intervention.
	StackTrace *RuntimeStackTrace `json:"stackTrace,omitempty"`
}

type AuditsInspectorIssueCode string

type AuditsInspectorIssueDetails struct {
	CookieIssueDetails                           *AuditsCookieIssueDetails                           `json:"cookieIssueDetails,omitempty"`
	MixedContentIssueDetails                     *AuditsMixedContentIssueDetails                     `json:"mixedContentIssueDetails,omitempty"`
	BlockedByResponseIssueDetails                *AuditsBlockedByResponseIssueDetails                `json:"blockedByResponseIssueDetails,omitempty"`
	HeavyAdIssueDetails                          *AuditsHeavyAdIssueDetails                          `json:"heavyAdIssueDetails,omitempty"`
	ContentSecurityPolicyIssueDetails            *AuditsContentSecurityPolicyIssueDetails            `json:"contentSecurityPolicyIssueDetails,omitempty"`
	SharedArrayBufferIssueDetails                *AuditsSharedArrayBufferIssueDetails                `json:"sharedArrayBufferIssueDetails,omitempty"`
	CorsIssueDetails                             *AuditsCorsIssueDetails                             `json:"corsIssueDetails,omitempty"`
	AttributionReportingIssueDetails             *AuditsAttributionReportingIssueDetails             `json:"attributionReportingIssueDetails,omitempty"`
	QuirksModeIssueDetails                       *AuditsQuirksModeIssueDetails                       `json:"quirksModeIssueDetails,omitempty"`
	PartitioningBlobURLIssueDetails              *AuditsPartitioningBlobURLIssueDetails              `json:"partitioningBlobURLIssueDetails,omitempty"`
	NavigatorUserAgentIssueDetails               *AuditsNavigatorUserAgentIssueDetails               `json:"navigatorUserAgentIssueDetails,omitempty"`
	GenericIssueDetails                          *AuditsGenericIssueDetails                          `json:"genericIssueDetails,omitempty"`
	DeprecationIssueDetails                      *AuditsDeprecationIssueDetails                      `json:"deprecationIssueDetails,omitempty"`
	ClientHintIssueDetails                       *AuditsClientHintIssueDetails                       `json:"clientHintIssueDetails,omitempty"`
	FederatedAuthRequestIssueDetails             *AuditsFederatedAuthRequestIssueDetails             `json:"federatedAuthRequestIssueDetails,omitempty"`
	BounceTrackingIssueDetails                   *AuditsBounceTrackingIssueDetails                   `json:"bounceTrackingIssueDetails,omitempty"`
	CookieDeprecationMetadataIssueDetails        *AuditsCookieDeprecationMetadataIssueDetails        `json:"cookieDeprecationMetadataIssueDetails,omitempty"`
	StylesheetLoadingIssueDetails                *AuditsStylesheetLoadingIssueDetails                `json:"stylesheetLoadingIssueDetails,omitempty"`
	PropertyRuleIssueDetails                     *AuditsPropertyRuleIssueDetails                     `json:"propertyRuleIssueDetails,omitempty"`
	FederatedAuthUserInfoRequestIssueDetails     *AuditsFederatedAuthUserInfoRequestIssueDetails     `json:"federatedAuthUserInfoRequestIssueDetails,omitempty"`
	SharedDictionaryIssueDetails                 *AuditsSharedDictionaryIssueDetails                 `json:"sharedDictionaryIssueDetails,omitempty"`
	ElementAccessibilityIssueDetails             *AuditsElementAccessibilityIssueDetails             `json:"elementAccessibilityIssueDetails,omitempty"`
	SriMessageSignatureIssueDetails              *AuditsSRIMessageSignatureIssueDetails              `json:"sriMessageSignatureIssueDetails,omitempty"`
	UnencodedDigestIssueDetails                  *AuditsUnencodedDigestIssueDetails                  `json:"unencodedDigestIssueDetails,omitempty"`
	ConnectionAllowlistIssueDetails              *AuditsConnectionAllowlistIssueDetails              `json:"connectionAllowlistIssueDetails,omitempty"`
	UserReidentificationIssueDetails             *AuditsUserReidentificationIssueDetails             `json:"userReidentificationIssueDetails,omitempty"`
	PermissionElementIssueDetails                *AuditsPermissionElementIssueDetails                `json:"permissionElementIssueDetails,omitempty"`
	PerformanceIssueDetails                      *AuditsPerformanceIssueDetails                      `json:"performanceIssueDetails,omitempty"`
	SelectivePermissionsInterventionIssueDetails *AuditsSelectivePermissionsInterventionIssueDetails `json:"selectivePermissionsInterventionIssueDetails,omitempty"`
}

type AuditsIssueID string

type AuditsInspectorIssue struct {
	Code    AuditsInspectorIssueCode    `json:"code"`
	Details AuditsInspectorIssueDetails `json:"details"`
	// A unique id for this issue. May be omitted if no other entity (e.g.
	IssueID *AuditsIssueID `json:"issueId,omitempty"`
}

type AuditsGetEncodedResponseParams struct {
	SessionID string `json:"-"`
	// Identifier of the network request to get content for.
	RequestID NetworkRequestID `json:"requestId"`
	// The encoding to use.
	Encoding string `json:"encoding"`
	// The quality of the encoding (0-1). (defaults to 1)
	Quality *float64 `json:"quality,omitempty"`
	// Whether to only return the size information (defaults to false).
	SizeOnly *bool `json:"sizeOnly,omitempty"`
}

type AuditsGetEncodedResponseResult struct {
	// The encoded body as a base64 string. Omitted if sizeOnly is true. (Encoded as a base64 string when passed over JSON)
	Body *string `json:"body,omitempty"`
	// Size before re-encoding.
	OriginalSize int `json:"originalSize"`
	// Size after re-encoding.
	EncodedSize int `json:"encodedSize"`
}

type AuditsDisableParams struct {
	SessionID string `json:"-"`
}

type AuditsDisableResult struct {
}

type AuditsEnableParams struct {
	SessionID string `json:"-"`
}

type AuditsEnableResult struct {
}

type AuditsCheckFormsIssuesParams struct {
	SessionID string `json:"-"`
}

type AuditsCheckFormsIssuesResult struct {
	FormIssues []AuditsGenericIssueDetails `json:"formIssues"`
}

type AuditsIssueAddedEvent struct {
	Issue AuditsInspectorIssue `json:"issue"`
}

type AutofillCreditCard struct {
	// 16-digit credit card number.
	Number string `json:"number"`
	// Name of the credit card owner.
	Name string `json:"name"`
	// 2-digit expiry month.
	ExpiryMonth string `json:"expiryMonth"`
	// 4-digit expiry year.
	ExpiryYear string `json:"expiryYear"`
	// 3-digit card verification code.
	Cvc string `json:"cvc"`
}

type AutofillAddressField struct {
	// address field name, for example GIVEN_NAME.
	Name string `json:"name"`
	// address field value, for example Jon Doe.
	Value string `json:"value"`
}

type AutofillAddressFields struct {
	Fields []AutofillAddressField `json:"fields"`
}

type AutofillAddress struct {
	// fields and values defining an address.
	Fields []AutofillAddressField `json:"fields"`
}

type AutofillAddressUI struct {
	// A two dimension array containing the representation of values from an address profile.
	AddressFields []AutofillAddressFields `json:"addressFields"`
}

type AutofillFillingStrategy string

type AutofillFilledField struct {
	// The type of the field, e.g text, password etc.
	HTMLType string `json:"htmlType"`
	// the html id
	ID string `json:"id"`
	// the html name
	Name string `json:"name"`
	// the field value
	Value string `json:"value"`
	// The actual field type, e.g FAMILY_NAME
	AutofillType string `json:"autofillType"`
	// The filling strategy
	FillingStrategy AutofillFillingStrategy `json:"fillingStrategy"`
	// The frame the field belongs to
	FrameID PageFrameID `json:"frameId"`
	// The form field's DOM node
	FieldID DOMBackendNodeID `json:"fieldId"`
}

type AutofillTriggerParams struct {
	SessionID string `json:"-"`
	// Identifies a field that serves as an anchor for autofill.
	FieldID DOMBackendNodeID `json:"fieldId"`
	// Identifies the frame that field belongs to.
	FrameID *PageFrameID `json:"frameId,omitempty"`
	// Credit card information to fill out the form. Credit card data is not saved.  Mutually exclusive with `address`.
	Card *AutofillCreditCard `json:"card,omitempty"`
	// Address to fill out the form. Address data is not saved. Mutually exclusive with `card`.
	Address *AutofillAddress `json:"address,omitempty"`
}

type AutofillTriggerResult struct {
}

type AutofillSetAddressesParams struct {
	SessionID string            `json:"-"`
	Addresses []AutofillAddress `json:"addresses"`
}

type AutofillSetAddressesResult struct {
}

type AutofillDisableParams struct {
	SessionID string `json:"-"`
}

type AutofillDisableResult struct {
}

type AutofillEnableParams struct {
	SessionID string `json:"-"`
}

type AutofillEnableResult struct {
}

type AutofillAddressFormFilledEvent struct {
	// Information about the fields that were filled
	FilledFields []AutofillFilledField `json:"filledFields"`
	// An UI representation of the address used to fill the form.
	AddressUI AutofillAddressUI `json:"addressUi"`
}

type BackgroundServiceServiceName string

type BackgroundServiceEventMetadata struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type BackgroundServiceBackgroundServiceEvent struct {
	// Timestamp of the event (in seconds).
	Timestamp NetworkTimeSinceEpoch `json:"timestamp"`
	// The origin this event belongs to.
	Origin string `json:"origin"`
	// The Service Worker ID that initiated the event.
	ServiceWorkerRegistrationID ServiceWorkerRegistrationID `json:"serviceWorkerRegistrationId"`
	// The Background Service this event belongs to.
	Service BackgroundServiceServiceName `json:"service"`
	// A description of the event.
	EventName string `json:"eventName"`
	// An identifier that groups related events together.
	InstanceID string `json:"instanceId"`
	// A list of event-specific information.
	EventMetadata []BackgroundServiceEventMetadata `json:"eventMetadata"`
	// Storage key this event belongs to.
	StorageKey string `json:"storageKey"`
}

type BackgroundServiceStartObservingParams struct {
	SessionID string                       `json:"-"`
	Service   BackgroundServiceServiceName `json:"service"`
}

type BackgroundServiceStartObservingResult struct {
}

type BackgroundServiceStopObservingParams struct {
	SessionID string                       `json:"-"`
	Service   BackgroundServiceServiceName `json:"service"`
}

type BackgroundServiceStopObservingResult struct {
}

type BackgroundServiceSetRecordingParams struct {
	SessionID    string                       `json:"-"`
	ShouldRecord bool                         `json:"shouldRecord"`
	Service      BackgroundServiceServiceName `json:"service"`
}

type BackgroundServiceSetRecordingResult struct {
}

type BackgroundServiceClearEventsParams struct {
	SessionID string                       `json:"-"`
	Service   BackgroundServiceServiceName `json:"service"`
}

type BackgroundServiceClearEventsResult struct {
}

type BackgroundServiceRecordingStateChangedEvent struct {
	IsRecording bool                         `json:"isRecording"`
	Service     BackgroundServiceServiceName `json:"service"`
}

type BackgroundServiceBackgroundServiceEventReceivedEvent struct {
	BackgroundServiceEvent BackgroundServiceBackgroundServiceEvent `json:"backgroundServiceEvent"`
}

type BluetoothEmulationCentralState string

type BluetoothEmulationGATTOperationType string

type BluetoothEmulationCharacteristicWriteType string

type BluetoothEmulationCharacteristicOperationType string

type BluetoothEmulationDescriptorOperationType string

type BluetoothEmulationManufacturerData struct {
	// Company identifier
	Key int `json:"key"`
	// Manufacturer-specific data (Encoded as a base64 string when passed over JSON)
	Data string `json:"data"`
}

type BluetoothEmulationScanRecord struct {
	Name  *string  `json:"name,omitempty"`
	Uuids []string `json:"uuids,omitempty"`
	// Stores the external appearance description of the device.
	Appearance *int `json:"appearance,omitempty"`
	// Stores the transmission power of a broadcasting device.
	TxPower *int `json:"txPower,omitempty"`
	// Key is the company identifier and the value is an array of bytes of
	ManufacturerData []BluetoothEmulationManufacturerData `json:"manufacturerData,omitempty"`
}

type BluetoothEmulationScanEntry struct {
	DeviceAddress string                       `json:"deviceAddress"`
	Rssi          int                          `json:"rssi"`
	ScanRecord    BluetoothEmulationScanRecord `json:"scanRecord"`
}

type BluetoothEmulationCharacteristicProperties struct {
	Broadcast                 *bool `json:"broadcast,omitempty"`
	Read                      *bool `json:"read,omitempty"`
	WriteWithoutResponse      *bool `json:"writeWithoutResponse,omitempty"`
	Write                     *bool `json:"write,omitempty"`
	Notify                    *bool `json:"notify,omitempty"`
	Indicate                  *bool `json:"indicate,omitempty"`
	AuthenticatedSignedWrites *bool `json:"authenticatedSignedWrites,omitempty"`
	ExtendedProperties        *bool `json:"extendedProperties,omitempty"`
}

type BluetoothEmulationEnableParams struct {
	SessionID string `json:"-"`
	// State of the simulated central.
	State BluetoothEmulationCentralState `json:"state"`
	// If the simulated central supports low-energy.
	LeSupported bool `json:"leSupported"`
}

type BluetoothEmulationEnableResult struct {
}

type BluetoothEmulationSetSimulatedCentralStateParams struct {
	SessionID string `json:"-"`
	// State of the simulated central.
	State BluetoothEmulationCentralState `json:"state"`
}

type BluetoothEmulationSetSimulatedCentralStateResult struct {
}

type BluetoothEmulationDisableParams struct {
	SessionID string `json:"-"`
}

type BluetoothEmulationDisableResult struct {
}

type BluetoothEmulationSimulatePreconnectedPeripheralParams struct {
	SessionID         string                               `json:"-"`
	Address           string                               `json:"address"`
	Name              string                               `json:"name"`
	ManufacturerData  []BluetoothEmulationManufacturerData `json:"manufacturerData"`
	KnownServiceUuids []string                             `json:"knownServiceUuids"`
}

type BluetoothEmulationSimulatePreconnectedPeripheralResult struct {
}

type BluetoothEmulationSimulateAdvertisementParams struct {
	SessionID string                      `json:"-"`
	Entry     BluetoothEmulationScanEntry `json:"entry"`
}

type BluetoothEmulationSimulateAdvertisementResult struct {
}

type BluetoothEmulationSimulateGATTOperationResponseParams struct {
	SessionID string                              `json:"-"`
	Address   string                              `json:"address"`
	Type      BluetoothEmulationGATTOperationType `json:"type"`
	Code      int                                 `json:"code"`
}

type BluetoothEmulationSimulateGATTOperationResponseResult struct {
}

type BluetoothEmulationSimulateCharacteristicOperationResponseParams struct {
	SessionID        string                                        `json:"-"`
	CharacteristicID string                                        `json:"characteristicId"`
	Type             BluetoothEmulationCharacteristicOperationType `json:"type"`
	Code             int                                           `json:"code"`
	Data             *string                                       `json:"data,omitempty"`
}

type BluetoothEmulationSimulateCharacteristicOperationResponseResult struct {
}

type BluetoothEmulationSimulateDescriptorOperationResponseParams struct {
	SessionID    string                                    `json:"-"`
	DescriptorID string                                    `json:"descriptorId"`
	Type         BluetoothEmulationDescriptorOperationType `json:"type"`
	Code         int                                       `json:"code"`
	Data         *string                                   `json:"data,omitempty"`
}

type BluetoothEmulationSimulateDescriptorOperationResponseResult struct {
}

type BluetoothEmulationAddServiceParams struct {
	SessionID   string `json:"-"`
	Address     string `json:"address"`
	ServiceUUID string `json:"serviceUuid"`
}

type BluetoothEmulationAddServiceResult struct {
	// An identifier that uniquely represents this service.
	ServiceID string `json:"serviceId"`
}

type BluetoothEmulationRemoveServiceParams struct {
	SessionID string `json:"-"`
	ServiceID string `json:"serviceId"`
}

type BluetoothEmulationRemoveServiceResult struct {
}

type BluetoothEmulationAddCharacteristicParams struct {
	SessionID          string                                     `json:"-"`
	ServiceID          string                                     `json:"serviceId"`
	CharacteristicUUID string                                     `json:"characteristicUuid"`
	Properties         BluetoothEmulationCharacteristicProperties `json:"properties"`
}

type BluetoothEmulationAddCharacteristicResult struct {
	// An identifier that uniquely represents this characteristic.
	CharacteristicID string `json:"characteristicId"`
}

type BluetoothEmulationRemoveCharacteristicParams struct {
	SessionID        string `json:"-"`
	CharacteristicID string `json:"characteristicId"`
}

type BluetoothEmulationRemoveCharacteristicResult struct {
}

type BluetoothEmulationAddDescriptorParams struct {
	SessionID        string `json:"-"`
	CharacteristicID string `json:"characteristicId"`
	DescriptorUUID   string `json:"descriptorUuid"`
}

type BluetoothEmulationAddDescriptorResult struct {
	// An identifier that uniquely represents this descriptor.
	DescriptorID string `json:"descriptorId"`
}

type BluetoothEmulationRemoveDescriptorParams struct {
	SessionID    string `json:"-"`
	DescriptorID string `json:"descriptorId"`
}

type BluetoothEmulationRemoveDescriptorResult struct {
}

type BluetoothEmulationSimulateGATTDisconnectionParams struct {
	SessionID string `json:"-"`
	Address   string `json:"address"`
}

type BluetoothEmulationSimulateGATTDisconnectionResult struct {
}

type BluetoothEmulationGattOperationReceivedEvent struct {
	Address string                              `json:"address"`
	Type    BluetoothEmulationGATTOperationType `json:"type"`
}

type BluetoothEmulationCharacteristicOperationReceivedEvent struct {
	CharacteristicID string                                        `json:"characteristicId"`
	Type             BluetoothEmulationCharacteristicOperationType `json:"type"`
	Data             *string                                       `json:"data,omitempty"`
	WriteType        *BluetoothEmulationCharacteristicWriteType    `json:"writeType,omitempty"`
}

type BluetoothEmulationDescriptorOperationReceivedEvent struct {
	DescriptorID string                                    `json:"descriptorId"`
	Type         BluetoothEmulationDescriptorOperationType `json:"type"`
	Data         *string                                   `json:"data,omitempty"`
}

type BrowserBrowserContextID string

type BrowserWindowID int

type BrowserWindowState string

type BrowserBounds struct {
	// The offset from the left edge of the screen to the window in pixels.
	Left *int `json:"left,omitempty"`
	// The offset from the top edge of the screen to the window in pixels.
	Top *int `json:"top,omitempty"`
	// The window width in pixels.
	Width *int `json:"width,omitempty"`
	// The window height in pixels.
	Height *int `json:"height,omitempty"`
	// The window state. Default to normal.
	WindowState *BrowserWindowState `json:"windowState,omitempty"`
}

type BrowserPermissionType string

type BrowserPermissionSetting string

type BrowserPermissionDescriptor struct {
	// Name of permission.
	Name string `json:"name"`
	// For "midi" permission, may also specify sysex control.
	Sysex *bool `json:"sysex,omitempty"`
	// For "push" permission, may specify userVisibleOnly.
	UserVisibleOnly *bool `json:"userVisibleOnly,omitempty"`
	// For "clipboard" permission, may specify allowWithoutSanitization.
	AllowWithoutSanitization *bool `json:"allowWithoutSanitization,omitempty"`
	// For "fullscreen" permission, must specify allowWithoutGesture:true.
	AllowWithoutGesture *bool `json:"allowWithoutGesture,omitempty"`
	// For "camera" permission, may specify panTiltZoom.
	PanTiltZoom *bool `json:"panTiltZoom,omitempty"`
}

type BrowserBrowserCommandID string

type BrowserBucket struct {
	// Minimum value (inclusive).
	Low int `json:"low"`
	// Maximum value (exclusive).
	High int `json:"high"`
	// Number of samples.
	Count int `json:"count"`
}

type BrowserHistogram struct {
	// Name.
	Name string `json:"name"`
	// Sum of sample values.
	Sum int `json:"sum"`
	// Total number of samples.
	Count int `json:"count"`
	// Buckets.
	Buckets []BrowserBucket `json:"buckets"`
}

type BrowserPrivacySandboxAPI string

type BrowserSetPermissionParams struct {
	SessionID string `json:"-"`
	// Descriptor of permission to override.
	Permission BrowserPermissionDescriptor `json:"permission"`
	// Setting of the permission.
	Setting BrowserPermissionSetting `json:"setting"`
	// Embedding origin the permission applies to, all origins if not specified.
	Origin *string `json:"origin,omitempty"`
	// Embedded origin the permission applies to. It is ignored unless the embedding origin is
	EmbeddedOrigin *string `json:"embeddedOrigin,omitempty"`
	// Context to override. When omitted, default browser context is used.
	BrowserContextID *BrowserBrowserContextID `json:"browserContextId,omitempty"`
}

type BrowserSetPermissionResult struct {
}

type BrowserGrantPermissionsParams struct {
	SessionID   string                  `json:"-"`
	Permissions []BrowserPermissionType `json:"permissions"`
	// Origin the permission applies to, all origins if not specified.
	Origin *string `json:"origin,omitempty"`
	// BrowserContext to override permissions. When omitted, default browser context is used.
	BrowserContextID *BrowserBrowserContextID `json:"browserContextId,omitempty"`
}

type BrowserGrantPermissionsResult struct {
}

type BrowserResetPermissionsParams struct {
	SessionID string `json:"-"`
	// BrowserContext to reset permissions. When omitted, default browser context is used.
	BrowserContextID *BrowserBrowserContextID `json:"browserContextId,omitempty"`
}

type BrowserResetPermissionsResult struct {
}

type BrowserSetDownloadBehaviorParams struct {
	SessionID string `json:"-"`
	// Whether to allow all or deny all download requests, or use default Chrome behavior if
	Behavior string `json:"behavior"`
	// BrowserContext to set download behavior. When omitted, default browser context is used.
	BrowserContextID *BrowserBrowserContextID `json:"browserContextId,omitempty"`
	// The default path to save downloaded files to. This is required if behavior is set to 'allow'
	DownloadPath *string `json:"downloadPath,omitempty"`
	// Whether to emit download events (defaults to false).
	EventsEnabled *bool `json:"eventsEnabled,omitempty"`
}

type BrowserSetDownloadBehaviorResult struct {
}

type BrowserCancelDownloadParams struct {
	SessionID string `json:"-"`
	// Global unique identifier of the download.
	Guid string `json:"guid"`
	// BrowserContext to perform the action in. When omitted, default browser context is used.
	BrowserContextID *BrowserBrowserContextID `json:"browserContextId,omitempty"`
}

type BrowserCancelDownloadResult struct {
}

type BrowserCloseParams struct {
	SessionID string `json:"-"`
}

type BrowserCloseResult struct {
}

type BrowserCrashParams struct {
	SessionID string `json:"-"`
}

type BrowserCrashResult struct {
}

type BrowserCrashGPUProcessParams struct {
	SessionID string `json:"-"`
}

type BrowserCrashGPUProcessResult struct {
}

type BrowserGetVersionParams struct {
	SessionID string `json:"-"`
}

type BrowserGetVersionResult struct {
	// Protocol version.
	ProtocolVersion string `json:"protocolVersion"`
	// Product name.
	Product string `json:"product"`
	// Product revision.
	Revision string `json:"revision"`
	// User-Agent.
	UserAgent string `json:"userAgent"`
	// V8 version.
	JSVersion string `json:"jsVersion"`
}

type BrowserGetBrowserCommandLineParams struct {
	SessionID string `json:"-"`
}

type BrowserGetBrowserCommandLineResult struct {
	// Commandline parameters
	Arguments []string `json:"arguments"`
}

type BrowserGetHistogramsParams struct {
	SessionID string `json:"-"`
	// Requested substring in name. Only histograms which have query as a
	Query *string `json:"query,omitempty"`
	// If true, retrieve delta since last delta call.
	Delta *bool `json:"delta,omitempty"`
}

type BrowserGetHistogramsResult struct {
	// Histograms.
	Histograms []BrowserHistogram `json:"histograms"`
}

type BrowserGetHistogramParams struct {
	SessionID string `json:"-"`
	// Requested histogram name.
	Name string `json:"name"`
	// If true, retrieve delta since last delta call.
	Delta *bool `json:"delta,omitempty"`
}

type BrowserGetHistogramResult struct {
	// Histogram.
	Histogram BrowserHistogram `json:"histogram"`
}

type BrowserGetWindowBoundsParams struct {
	SessionID string `json:"-"`
	// Browser window id.
	WindowID BrowserWindowID `json:"windowId"`
}

type BrowserGetWindowBoundsResult struct {
	// Bounds information of the window. When window state is 'minimized', the restored window
	Bounds BrowserBounds `json:"bounds"`
}

type BrowserGetWindowForTargetParams struct {
	SessionID string `json:"-"`
	// Devtools agent host id. If called as a part of the session, associated targetId is used.
	TargetID *TargetTargetID `json:"targetId,omitempty"`
}

type BrowserGetWindowForTargetResult struct {
	// Browser window id.
	WindowID BrowserWindowID `json:"windowId"`
	// Bounds information of the window. When window state is 'minimized', the restored window
	Bounds BrowserBounds `json:"bounds"`
}

type BrowserSetWindowBoundsParams struct {
	SessionID string `json:"-"`
	// Browser window id.
	WindowID BrowserWindowID `json:"windowId"`
	// New window bounds. The 'minimized', 'maximized' and 'fullscreen' states cannot be combined
	Bounds BrowserBounds `json:"bounds"`
}

type BrowserSetWindowBoundsResult struct {
}

type BrowserSetContentsSizeParams struct {
	SessionID string `json:"-"`
	// Browser window id.
	WindowID BrowserWindowID `json:"windowId"`
	// The window contents width in DIP. Assumes current width if omitted.
	Width *int `json:"width,omitempty"`
	// The window contents height in DIP. Assumes current height if omitted.
	Height *int `json:"height,omitempty"`
}

type BrowserSetContentsSizeResult struct {
}

type BrowserSetDockTileParams struct {
	SessionID  string  `json:"-"`
	BadgeLabel *string `json:"badgeLabel,omitempty"`
	// Png encoded image. (Encoded as a base64 string when passed over JSON)
	Image *string `json:"image,omitempty"`
}

type BrowserSetDockTileResult struct {
}

type BrowserExecuteBrowserCommandParams struct {
	SessionID string                  `json:"-"`
	CommandID BrowserBrowserCommandID `json:"commandId"`
}

type BrowserExecuteBrowserCommandResult struct {
}

type BrowserAddPrivacySandboxEnrollmentOverrideParams struct {
	SessionID string `json:"-"`
	URL       string `json:"url"`
}

type BrowserAddPrivacySandboxEnrollmentOverrideResult struct {
}

type BrowserAddPrivacySandboxCoordinatorKeyConfigParams struct {
	SessionID         string                   `json:"-"`
	API               BrowserPrivacySandboxAPI `json:"api"`
	CoordinatorOrigin string                   `json:"coordinatorOrigin"`
	KeyConfig         string                   `json:"keyConfig"`
	// BrowserContext to perform the action in. When omitted, default browser
	BrowserContextID *BrowserBrowserContextID `json:"browserContextId,omitempty"`
}

type BrowserAddPrivacySandboxCoordinatorKeyConfigResult struct {
}

type BrowserDownloadWillBeginEvent struct {
	// Id of the frame that caused the download to begin.
	FrameID PageFrameID `json:"frameId"`
	// Global unique identifier of the download.
	Guid string `json:"guid"`
	// URL of the resource being downloaded.
	URL string `json:"url"`
	// Suggested file name of the resource (the actual name of the file saved on disk may differ).
	SuggestedFilename string `json:"suggestedFilename"`
}

type BrowserDownloadProgressEvent struct {
	// Global unique identifier of the download.
	Guid string `json:"guid"`
	// Total expected bytes to download.
	TotalBytes float64 `json:"totalBytes"`
	// Total bytes received.
	ReceivedBytes float64 `json:"receivedBytes"`
	// Download status.
	State string `json:"state"`
	// If download is "completed", provides the path of the downloaded file.
	FilePath *string `json:"filePath,omitempty"`
}

type CSSStyleSheetOrigin string

type CSSPseudoElementMatches struct {
	// Pseudo element type.
	PseudoType DOMPseudoType `json:"pseudoType"`
	// Pseudo element custom ident.
	PseudoIdentifier *string `json:"pseudoIdentifier,omitempty"`
	// Matches of CSS rules applicable to the pseudo style.
	Matches []CSSRuleMatch `json:"matches"`
}

type CSSCSSAnimationStyle struct {
	// The name of the animation.
	Name *string `json:"name,omitempty"`
	// The style coming from the animation.
	Style CSSCSSStyle `json:"style"`
}

type CSSInheritedStyleEntry struct {
	// The ancestor node's inline style, if any, in the style inheritance chain.
	InlineStyle *CSSCSSStyle `json:"inlineStyle,omitempty"`
	// Matches of CSS rules matching the ancestor node in the style inheritance chain.
	MatchedCSSRules []CSSRuleMatch `json:"matchedCSSRules"`
}

type CSSInheritedAnimatedStyleEntry struct {
	// Styles coming from the animations of the ancestor, if any, in the style inheritance chain.
	AnimationStyles []CSSCSSAnimationStyle `json:"animationStyles,omitempty"`
	// The style coming from the transitions of the ancestor, if any, in the style inheritance chain.
	TransitionsStyle *CSSCSSStyle `json:"transitionsStyle,omitempty"`
}

type CSSInheritedPseudoElementMatches struct {
	// Matches of pseudo styles from the pseudos of an ancestor node.
	PseudoElements []CSSPseudoElementMatches `json:"pseudoElements"`
}

type CSSRuleMatch struct {
	// CSS rule in the match.
	Rule CSSCSSRule `json:"rule"`
	// Matching selector indices in the rule's selectorList selectors (0-based).
	MatchingSelectors []int `json:"matchingSelectors"`
}

type CSSValue struct {
	// Value text.
	Text string `json:"text"`
	// Value range in the underlying resource (if available).
	Range *CSSSourceRange `json:"range,omitempty"`
	// Specificity of the selector.
	Specificity *CSSSpecificity `json:"specificity,omitempty"`
}

type CSSSpecificity struct {
	// The a component, which represents the number of ID selectors.
	A int `json:"a"`
	// The b component, which represents the number of class selectors, attributes selectors, and
	B int `json:"b"`
	// The c component, which represents the number of type selectors and pseudo-elements.
	C int `json:"c"`
}

type CSSSelectorList struct {
	// Selectors in the list.
	Selectors []CSSValue `json:"selectors"`
	// Rule selector text.
	Text string `json:"text"`
}

type CSSCSSStyleSheetHeader struct {
	// The stylesheet identifier.
	StyleSheetID DOMStyleSheetID `json:"styleSheetId"`
	// Owner frame identifier.
	FrameID PageFrameID `json:"frameId"`
	// Stylesheet resource URL. Empty if this is a constructed stylesheet created using
	SourceURL string `json:"sourceURL"`
	// URL of source map associated with the stylesheet (if any).
	SourceMapURL *string `json:"sourceMapURL,omitempty"`
	// Stylesheet origin.
	Origin CSSStyleSheetOrigin `json:"origin"`
	// Stylesheet title.
	Title string `json:"title"`
	// The backend id for the owner node of the stylesheet.
	OwnerNode *DOMBackendNodeID `json:"ownerNode,omitempty"`
	// Denotes whether the stylesheet is disabled.
	Disabled bool `json:"disabled"`
	// Whether the sourceURL field value comes from the sourceURL comment.
	HasSourceURL *bool `json:"hasSourceURL,omitempty"`
	// Whether this stylesheet is created for STYLE tag by parser. This flag is not set for
	IsInline bool `json:"isInline"`
	// Whether this stylesheet is mutable. Inline stylesheets become mutable
	IsMutable bool `json:"isMutable"`
	// True if this stylesheet is created through new CSSStyleSheet() or imported as a
	IsConstructed bool `json:"isConstructed"`
	// Line offset of the stylesheet within the resource (zero based).
	StartLine float64 `json:"startLine"`
	// Column offset of the stylesheet within the resource (zero based).
	StartColumn float64 `json:"startColumn"`
	// Size of the content (in characters).
	Length float64 `json:"length"`
	// Line offset of the end of the stylesheet within the resource (zero based).
	EndLine float64 `json:"endLine"`
	// Column offset of the end of the stylesheet within the resource (zero based).
	EndColumn float64 `json:"endColumn"`
	// If the style sheet was loaded from a network resource, this indicates when the resource failed to load
	LoadingFailed *bool `json:"loadingFailed,omitempty"`
}

type CSSCSSRule struct {
	// The css style sheet identifier (absent for user agent stylesheet and user-specified
	StyleSheetID *DOMStyleSheetID `json:"styleSheetId,omitempty"`
	// Rule selector data.
	SelectorList CSSSelectorList `json:"selectorList"`
	// Array of selectors from ancestor style rules, sorted by distance from the current rule.
	NestingSelectors []string `json:"nestingSelectors,omitempty"`
	// Parent stylesheet's origin.
	Origin CSSStyleSheetOrigin `json:"origin"`
	// Associated style declaration.
	Style CSSCSSStyle `json:"style"`
	// The BackendNodeId of the DOM node that constitutes the origin tree scope of this rule.
	OriginTreeScopeNodeID *DOMBackendNodeID `json:"originTreeScopeNodeId,omitempty"`
	// Media list array (for rules involving media queries). The array enumerates media queries
	Media []CSSCSSMedia `json:"media,omitempty"`
	// Container query list array (for rules involving container queries).
	ContainerQueries []CSSCSSContainerQuery `json:"containerQueries,omitempty"`
	// @supports CSS at-rule array.
	Supports []CSSCSSSupports `json:"supports,omitempty"`
	// Cascade layer array. Contains the layer hierarchy that this rule belongs to starting
	Layers []CSSCSSLayer `json:"layers,omitempty"`
	// @scope CSS at-rule array.
	Scopes []CSSCSSScope `json:"scopes,omitempty"`
	// The array keeps the types of ancestor CSSRules from the innermost going outwards.
	RuleTypes []CSSCSSRuleType `json:"ruleTypes,omitempty"`
	// @starting-style CSS at-rule array.
	StartingStyles []CSSCSSStartingStyle `json:"startingStyles,omitempty"`
	// @navigation CSS at-rule array.
	Navigations []CSSCSSNavigation `json:"navigations,omitempty"`
}

type CSSCSSRuleType string

type CSSRuleUsage struct {
	// The css style sheet identifier (absent for user agent stylesheet and user-specified
	StyleSheetID DOMStyleSheetID `json:"styleSheetId"`
	// Offset of the start of the rule (including selector) from the beginning of the stylesheet.
	StartOffset float64 `json:"startOffset"`
	// Offset of the end of the rule body from the beginning of the stylesheet.
	EndOffset float64 `json:"endOffset"`
	// Indicates whether the rule was actually used by some element in the page.
	Used bool `json:"used"`
}

type CSSSourceRange struct {
	// Start line of range.
	StartLine int `json:"startLine"`
	// Start column of range (inclusive).
	StartColumn int `json:"startColumn"`
	// End line of range
	EndLine int `json:"endLine"`
	// End column of range (exclusive).
	EndColumn int `json:"endColumn"`
}

type CSSShorthandEntry struct {
	// Shorthand name.
	Name string `json:"name"`
	// Shorthand value.
	Value string `json:"value"`
	// Whether the property has "!important" annotation (implies `false` if absent).
	Important *bool `json:"important,omitempty"`
}

type CSSCSSComputedStyleProperty struct {
	// Computed style property name.
	Name string `json:"name"`
	// Computed style property value.
	Value string `json:"value"`
}

type CSSComputedStyleExtraFields struct {
	// Returns whether or not this node is being rendered with base appearance,
	IsAppearanceBase bool `json:"isAppearanceBase"`
}

type CSSCSSStyle struct {
	// The css style sheet identifier (absent for user agent stylesheet and user-specified
	StyleSheetID *DOMStyleSheetID `json:"styleSheetId,omitempty"`
	// CSS properties in the style.
	CSSProperties []CSSCSSProperty `json:"cssProperties"`
	// Computed values for all shorthands found in the style.
	ShorthandEntries []CSSShorthandEntry `json:"shorthandEntries"`
	// Style declaration text (if available).
	CSSText *string `json:"cssText,omitempty"`
	// Style declaration range in the enclosing stylesheet (if available).
	Range *CSSSourceRange `json:"range,omitempty"`
}

type CSSCSSProperty struct {
	// The property name.
	Name string `json:"name"`
	// The property value.
	Value string `json:"value"`
	// Whether the property has "!important" annotation (implies `false` if absent).
	Important *bool `json:"important,omitempty"`
	// Whether the property is implicit (implies `false` if absent).
	Implicit *bool `json:"implicit,omitempty"`
	// The full property text as specified in the style.
	Text *string `json:"text,omitempty"`
	// Whether the property is understood by the browser (implies `true` if absent).
	ParsedOk *bool `json:"parsedOk,omitempty"`
	// Whether the property is disabled by the user (present for source-based properties only).
	Disabled *bool `json:"disabled,omitempty"`
	// The entire property range in the enclosing style declaration (if available).
	Range *CSSSourceRange `json:"range,omitempty"`
	// Parsed longhand components of this property if it is a shorthand.
	LonghandProperties []CSSCSSProperty `json:"longhandProperties,omitempty"`
}

type CSSCSSMedia struct {
	// Media query text.
	Text string `json:"text"`
	// Source of the media query: "mediaRule" if specified by a @media rule, "importRule" if
	Source string `json:"source"`
	// URL of the document containing the media query description.
	SourceURL *string `json:"sourceURL,omitempty"`
	// The associated rule (@media or @import) header range in the enclosing stylesheet (if
	Range *CSSSourceRange `json:"range,omitempty"`
	// Identifier of the stylesheet containing this object (if exists).
	StyleSheetID *DOMStyleSheetID `json:"styleSheetId,omitempty"`
	// Array of media queries.
	MediaList []CSSMediaQuery `json:"mediaList,omitempty"`
}

type CSSMediaQuery struct {
	// Array of media query expressions.
	Expressions []CSSMediaQueryExpression `json:"expressions"`
	// Whether the media query condition is satisfied.
	Active bool `json:"active"`
}

type CSSMediaQueryExpression struct {
	// Media query expression value.
	Value float64 `json:"value"`
	// Media query expression units.
	Unit string `json:"unit"`
	// Media query expression feature.
	Feature string `json:"feature"`
	// The associated range of the value text in the enclosing stylesheet (if available).
	ValueRange *CSSSourceRange `json:"valueRange,omitempty"`
	// Computed length of media query expression (if applicable).
	ComputedLength *float64 `json:"computedLength,omitempty"`
}

type CSSCSSContainerQuery struct {
	// Container query text.
	Text string `json:"text"`
	// The associated rule header range in the enclosing stylesheet (if
	Range *CSSSourceRange `json:"range,omitempty"`
	// Identifier of the stylesheet containing this object (if exists).
	StyleSheetID *DOMStyleSheetID `json:"styleSheetId,omitempty"`
	// Optional name for the container.
	Name *string `json:"name,omitempty"`
	// Optional physical axes queried for the container.
	PhysicalAxes *DOMPhysicalAxes `json:"physicalAxes,omitempty"`
	// Optional logical axes queried for the container.
	LogicalAxes *DOMLogicalAxes `json:"logicalAxes,omitempty"`
	// true if the query contains scroll-state() queries.
	QueriesScrollState *bool `json:"queriesScrollState,omitempty"`
	// true if the query contains anchored() queries.
	QueriesAnchored *bool `json:"queriesAnchored,omitempty"`
}

type CSSCSSSupports struct {
	// Supports rule text.
	Text string `json:"text"`
	// Whether the supports condition is satisfied.
	Active bool `json:"active"`
	// The associated rule header range in the enclosing stylesheet (if
	Range *CSSSourceRange `json:"range,omitempty"`
	// Identifier of the stylesheet containing this object (if exists).
	StyleSheetID *DOMStyleSheetID `json:"styleSheetId,omitempty"`
}

type CSSCSSNavigation struct {
	// Navigation rule text.
	Text string `json:"text"`
	// Whether the navigation condition is satisfied.
	Active *bool `json:"active,omitempty"`
	// The associated rule header range in the enclosing stylesheet (if
	Range *CSSSourceRange `json:"range,omitempty"`
	// Identifier of the stylesheet containing this object (if exists).
	StyleSheetID *DOMStyleSheetID `json:"styleSheetId,omitempty"`
}

type CSSCSSScope struct {
	// Scope rule text.
	Text string `json:"text"`
	// The associated rule header range in the enclosing stylesheet (if
	Range *CSSSourceRange `json:"range,omitempty"`
	// Identifier of the stylesheet containing this object (if exists).
	StyleSheetID *DOMStyleSheetID `json:"styleSheetId,omitempty"`
}

type CSSCSSLayer struct {
	// Layer name.
	Text string `json:"text"`
	// The associated rule header range in the enclosing stylesheet (if
	Range *CSSSourceRange `json:"range,omitempty"`
	// Identifier of the stylesheet containing this object (if exists).
	StyleSheetID *DOMStyleSheetID `json:"styleSheetId,omitempty"`
}

type CSSCSSStartingStyle struct {
	// The associated rule header range in the enclosing stylesheet (if
	Range *CSSSourceRange `json:"range,omitempty"`
	// Identifier of the stylesheet containing this object (if exists).
	StyleSheetID *DOMStyleSheetID `json:"styleSheetId,omitempty"`
}

type CSSCSSLayerData struct {
	// Layer name.
	Name string `json:"name"`
	// Direct sub-layers
	SubLayers []CSSCSSLayerData `json:"subLayers,omitempty"`
	// Layer order. The order determines the order of the layer in the cascade order.
	Order float64 `json:"order"`
}

type CSSPlatformFontUsage struct {
	// Font's family name reported by platform.
	FamilyName string `json:"familyName"`
	// Font's PostScript name reported by platform.
	PostScriptName string `json:"postScriptName"`
	// Indicates if the font was downloaded or resolved locally.
	IsCustomFont bool `json:"isCustomFont"`
	// Amount of glyphs that were rendered with this font.
	GlyphCount float64 `json:"glyphCount"`
}

type CSSFontVariationAxis struct {
	// The font-variation-setting tag (a.k.a. "axis tag").
	Tag string `json:"tag"`
	// Human-readable variation name in the default language (normally, "en").
	Name string `json:"name"`
	// The minimum value (inclusive) the font supports for this tag.
	MinValue float64 `json:"minValue"`
	// The maximum value (inclusive) the font supports for this tag.
	MaxValue float64 `json:"maxValue"`
	// The default value.
	DefaultValue float64 `json:"defaultValue"`
}

type CSSFontFace struct {
	// The font-family.
	FontFamily string `json:"fontFamily"`
	// The font-style.
	FontStyle string `json:"fontStyle"`
	// The font-variant.
	FontVariant string `json:"fontVariant"`
	// The font-weight.
	FontWeight string `json:"fontWeight"`
	// The font-stretch.
	FontStretch string `json:"fontStretch"`
	// The font-display.
	FontDisplay string `json:"fontDisplay"`
	// The unicode-range.
	UnicodeRange string `json:"unicodeRange"`
	// The src.
	Src string `json:"src"`
	// The resolved platform font family
	PlatformFontFamily string `json:"platformFontFamily"`
	// Available variation settings (a.k.a. "axes").
	FontVariationAxes []CSSFontVariationAxis `json:"fontVariationAxes,omitempty"`
}

type CSSCSSTryRule struct {
	// The css style sheet identifier (absent for user agent stylesheet and user-specified
	StyleSheetID *DOMStyleSheetID `json:"styleSheetId,omitempty"`
	// Parent stylesheet's origin.
	Origin CSSStyleSheetOrigin `json:"origin"`
	// Associated style declaration.
	Style CSSCSSStyle `json:"style"`
}

type CSSCSSPositionTryRule struct {
	// The prelude dashed-ident name
	Name CSSValue `json:"name"`
	// The css style sheet identifier (absent for user agent stylesheet and user-specified
	StyleSheetID *DOMStyleSheetID `json:"styleSheetId,omitempty"`
	// Parent stylesheet's origin.
	Origin CSSStyleSheetOrigin `json:"origin"`
	// Associated style declaration.
	Style  CSSCSSStyle `json:"style"`
	Active bool        `json:"active"`
}

type CSSCSSKeyframesRule struct {
	// Animation name.
	AnimationName CSSValue `json:"animationName"`
	// List of keyframes.
	Keyframes []CSSCSSKeyframeRule `json:"keyframes"`
}

type CSSCSSPropertyRegistration struct {
	PropertyName string    `json:"propertyName"`
	InitialValue *CSSValue `json:"initialValue,omitempty"`
	Inherits     bool      `json:"inherits"`
	Syntax       string    `json:"syntax"`
}

type CSSCSSAtRule struct {
	// Type of at-rule.
	Type string `json:"type"`
	// Subsection of font-feature-values, if this is a subsection.
	Subsection *string `json:"subsection,omitempty"`
	// LINT.ThenChange(//third_party/blink/renderer/core/inspector/inspector_style_sheet.cc:FontVariantAlternatesFeatureType,//third_party/blink/renderer/core/inspector/inspector_css_agen
	Name *CSSValue `json:"name,omitempty"`
	// The css style sheet identifier (absent for user agent stylesheet and user-specified
	StyleSheetID *DOMStyleSheetID `json:"styleSheetId,omitempty"`
	// Parent stylesheet's origin.
	Origin CSSStyleSheetOrigin `json:"origin"`
	// Associated style declaration.
	Style CSSCSSStyle `json:"style"`
}

type CSSCSSPropertyRule struct {
	// The css style sheet identifier (absent for user agent stylesheet and user-specified
	StyleSheetID *DOMStyleSheetID `json:"styleSheetId,omitempty"`
	// Parent stylesheet's origin.
	Origin CSSStyleSheetOrigin `json:"origin"`
	// Associated property name.
	PropertyName CSSValue `json:"propertyName"`
	// Associated style declaration.
	Style CSSCSSStyle `json:"style"`
}

type CSSCSSFunctionParameter struct {
	// The parameter name.
	Name string `json:"name"`
	// The parameter type.
	Type string `json:"type"`
}

type CSSCSSFunctionConditionNode struct {
	// Media query for this conditional block. Only one type of condition should be set.
	Media *CSSCSSMedia `json:"media,omitempty"`
	// Container query for this conditional block. Only one type of condition should be set.
	ContainerQueries *CSSCSSContainerQuery `json:"containerQueries,omitempty"`
	// @supports CSS at-rule condition. Only one type of condition should be set.
	Supports *CSSCSSSupports `json:"supports,omitempty"`
	// @navigation condition. Only one type of condition should be set.
	Navigation *CSSCSSNavigation `json:"navigation,omitempty"`
	// Block body.
	Children []CSSCSSFunctionNode `json:"children"`
	// The condition text.
	ConditionText string `json:"conditionText"`
}

type CSSCSSFunctionNode struct {
	// A conditional block. If set, style should not be set.
	Condition *CSSCSSFunctionConditionNode `json:"condition,omitempty"`
	// Values set by this node. If set, condition should not be set.
	Style *CSSCSSStyle `json:"style,omitempty"`
}

type CSSCSSFunctionRule struct {
	// Name of the function.
	Name CSSValue `json:"name"`
	// The css style sheet identifier (absent for user agent stylesheet and user-specified
	StyleSheetID *DOMStyleSheetID `json:"styleSheetId,omitempty"`
	// Parent stylesheet's origin.
	Origin CSSStyleSheetOrigin `json:"origin"`
	// List of parameters.
	Parameters []CSSCSSFunctionParameter `json:"parameters"`
	// Function body.
	Children []CSSCSSFunctionNode `json:"children"`
}

type CSSCSSKeyframeRule struct {
	// The css style sheet identifier (absent for user agent stylesheet and user-specified
	StyleSheetID *DOMStyleSheetID `json:"styleSheetId,omitempty"`
	// Parent stylesheet's origin.
	Origin CSSStyleSheetOrigin `json:"origin"`
	// Associated key text.
	KeyText CSSValue `json:"keyText"`
	// Associated style declaration.
	Style CSSCSSStyle `json:"style"`
}

type CSSStyleDeclarationEdit struct {
	// The css style sheet identifier.
	StyleSheetID DOMStyleSheetID `json:"styleSheetId"`
	// The range of the style text in the enclosing stylesheet.
	Range CSSSourceRange `json:"range"`
	// New style text.
	Text string `json:"text"`
}

type CSSAddRuleParams struct {
	SessionID string `json:"-"`
	// The css style sheet identifier where a new rule should be inserted.
	StyleSheetID DOMStyleSheetID `json:"styleSheetId"`
	// The text of a new rule.
	RuleText string `json:"ruleText"`
	// Text position of a new rule in the target style sheet.
	Location CSSSourceRange `json:"location"`
	// NodeId for the DOM node in whose context custom property declarations for registered properties should be
	NodeForPropertySyntaxValidation *DOMNodeID `json:"nodeForPropertySyntaxValidation,omitempty"`
}

type CSSAddRuleResult struct {
	// The newly created rule.
	Rule CSSCSSRule `json:"rule"`
}

type CSSCollectClassNamesParams struct {
	SessionID    string          `json:"-"`
	StyleSheetID DOMStyleSheetID `json:"styleSheetId"`
}

type CSSCollectClassNamesResult struct {
	// Class name list.
	ClassNames []string `json:"classNames"`
}

type CSSCreateStyleSheetParams struct {
	SessionID string `json:"-"`
	// Identifier of the frame where "via-inspector" stylesheet should be created.
	FrameID PageFrameID `json:"frameId"`
	// If true, creates a new stylesheet for every call. If false,
	Force *bool `json:"force,omitempty"`
}

type CSSCreateStyleSheetResult struct {
	// Identifier of the created "via-inspector" stylesheet.
	StyleSheetID DOMStyleSheetID `json:"styleSheetId"`
}

type CSSDisableParams struct {
	SessionID string `json:"-"`
}

type CSSDisableResult struct {
}

type CSSEnableParams struct {
	SessionID string `json:"-"`
}

type CSSEnableResult struct {
}

type CSSForcePseudoStateParams struct {
	SessionID string `json:"-"`
	// The element id for which to force the pseudo state.
	NodeID DOMNodeID `json:"nodeId"`
	// Element pseudo classes to force when computing the element's style.
	ForcedPseudoClasses []string `json:"forcedPseudoClasses"`
}

type CSSForcePseudoStateResult struct {
}

type CSSForceStartingStyleParams struct {
	SessionID string `json:"-"`
	// The element id for which to force the starting-style state.
	NodeID DOMNodeID `json:"nodeId"`
	// Boolean indicating if this is on or off.
	Forced bool `json:"forced"`
}

type CSSForceStartingStyleResult struct {
}

type CSSGetBackgroundColorsParams struct {
	SessionID string `json:"-"`
	// Id of the node to get background colors for.
	NodeID DOMNodeID `json:"nodeId"`
}

type CSSGetBackgroundColorsResult struct {
	// The range of background colors behind this element, if it contains any visible text. If no
	BackgroundColors []string `json:"backgroundColors,omitempty"`
	// The computed font size for this node, as a CSS computed value string (e.g. '12px').
	ComputedFontSize *string `json:"computedFontSize,omitempty"`
	// The computed font weight for this node, as a CSS computed value string (e.g. 'normal' or
	ComputedFontWeight *string `json:"computedFontWeight,omitempty"`
}

type CSSGetComputedStyleForNodeParams struct {
	SessionID string    `json:"-"`
	NodeID    DOMNodeID `json:"nodeId"`
}

type CSSGetComputedStyleForNodeResult struct {
	// Computed style for the specified DOM node.
	ComputedStyle []CSSCSSComputedStyleProperty `json:"computedStyle"`
	// A list of non-standard "extra fields" which blink stores alongside each
	ExtraFields CSSComputedStyleExtraFields `json:"extraFields"`
}

type CSSResolveValuesParams struct {
	SessionID string `json:"-"`
	// Cascade-dependent keywords (revert/revert-layer) do not work.
	Values []string `json:"values"`
	// Id of the node in whose context the expression is evaluated
	NodeID DOMNodeID `json:"nodeId"`
	// Only longhands and custom property names are accepted.
	PropertyName *string `json:"propertyName,omitempty"`
	// Pseudo element type, only works for pseudo elements that generate
	PseudoType *DOMPseudoType `json:"pseudoType,omitempty"`
	// Pseudo element custom ident.
	PseudoIdentifier *string `json:"pseudoIdentifier,omitempty"`
}

type CSSResolveValuesResult struct {
	Results []string `json:"results"`
}

type CSSGetLonghandPropertiesParams struct {
	SessionID     string `json:"-"`
	ShorthandName string `json:"shorthandName"`
	Value         string `json:"value"`
}

type CSSGetLonghandPropertiesResult struct {
	LonghandProperties []CSSCSSProperty `json:"longhandProperties"`
}

type CSSGetInlineStylesForNodeParams struct {
	SessionID string    `json:"-"`
	NodeID    DOMNodeID `json:"nodeId"`
}

type CSSGetInlineStylesForNodeResult struct {
	// Inline style for the specified DOM node.
	InlineStyle *CSSCSSStyle `json:"inlineStyle,omitempty"`
	// Attribute-defined element style (e.g. resulting from "width=20 height=100%").
	AttributesStyle *CSSCSSStyle `json:"attributesStyle,omitempty"`
}

type CSSGetAnimatedStylesForNodeParams struct {
	SessionID string    `json:"-"`
	NodeID    DOMNodeID `json:"nodeId"`
}

type CSSGetAnimatedStylesForNodeResult struct {
	// Styles coming from animations.
	AnimationStyles []CSSCSSAnimationStyle `json:"animationStyles,omitempty"`
	// Style coming from transitions.
	TransitionsStyle *CSSCSSStyle `json:"transitionsStyle,omitempty"`
	// Inherited style entries for animationsStyle and transitionsStyle from
	Inherited []CSSInheritedAnimatedStyleEntry `json:"inherited,omitempty"`
}

type CSSGetMatchedStylesForNodeParams struct {
	SessionID string    `json:"-"`
	NodeID    DOMNodeID `json:"nodeId"`
}

type CSSGetMatchedStylesForNodeResult struct {
	// Inline style for the specified DOM node.
	InlineStyle *CSSCSSStyle `json:"inlineStyle,omitempty"`
	// Attribute-defined element style (e.g. resulting from "width=20 height=100%").
	AttributesStyle *CSSCSSStyle `json:"attributesStyle,omitempty"`
	// CSS rules matching this node, from all applicable stylesheets.
	MatchedCSSRules []CSSRuleMatch `json:"matchedCSSRules,omitempty"`
	// Pseudo style matches for this node.
	PseudoElements []CSSPseudoElementMatches `json:"pseudoElements,omitempty"`
	// A chain of inherited styles (from the immediate node parent up to the DOM tree root).
	Inherited []CSSInheritedStyleEntry `json:"inherited,omitempty"`
	// A chain of inherited pseudo element styles (from the immediate node parent up to the DOM tree root).
	InheritedPseudoElements []CSSInheritedPseudoElementMatches `json:"inheritedPseudoElements,omitempty"`
	// A list of CSS keyframed animations matching this node.
	CSSKeyframesRules []CSSCSSKeyframesRule `json:"cssKeyframesRules,omitempty"`
	// A list of CSS @position-try rules matching this node, based on the position-try-fallbacks property.
	CSSPositionTryRules []CSSCSSPositionTryRule `json:"cssPositionTryRules,omitempty"`
	// Index of the active fallback in the applied position-try-fallback property,
	ActivePositionFallbackIndex *int `json:"activePositionFallbackIndex,omitempty"`
	// A list of CSS at-property rules matching this node.
	CSSPropertyRules []CSSCSSPropertyRule `json:"cssPropertyRules,omitempty"`
	// A list of CSS property registrations matching this node.
	CSSPropertyRegistrations []CSSCSSPropertyRegistration `json:"cssPropertyRegistrations,omitempty"`
	// A list of simple @rules matching this node or its pseudo-elements.
	CSSAtRules []CSSCSSAtRule `json:"cssAtRules,omitempty"`
	// Id of the first parent element that does not have display: contents.
	ParentLayoutNodeID *DOMNodeID `json:"parentLayoutNodeId,omitempty"`
	// A list of CSS at-function rules referenced by styles of this node.
	CSSFunctionRules []CSSCSSFunctionRule `json:"cssFunctionRules,omitempty"`
}

type CSSGetEnvironmentVariablesParams struct {
	SessionID string `json:"-"`
}

type CSSGetEnvironmentVariablesResult struct {
	EnvironmentVariables map[string]any `json:"environmentVariables"`
}

type CSSGetMediaQueriesParams struct {
	SessionID string `json:"-"`
}

type CSSGetMediaQueriesResult struct {
	Medias []CSSCSSMedia `json:"medias"`
}

type CSSGetPlatformFontsForNodeParams struct {
	SessionID string    `json:"-"`
	NodeID    DOMNodeID `json:"nodeId"`
}

type CSSGetPlatformFontsForNodeResult struct {
	// Usage statistics for every employed platform font.
	Fonts []CSSPlatformFontUsage `json:"fonts"`
}

type CSSGetStyleSheetTextParams struct {
	SessionID    string          `json:"-"`
	StyleSheetID DOMStyleSheetID `json:"styleSheetId"`
}

type CSSGetStyleSheetTextResult struct {
	// The stylesheet text.
	Text string `json:"text"`
}

type CSSGetLayersForNodeParams struct {
	SessionID string    `json:"-"`
	NodeID    DOMNodeID `json:"nodeId"`
}

type CSSGetLayersForNodeResult struct {
	RootLayer CSSCSSLayerData `json:"rootLayer"`
}

type CSSGetLocationForSelectorParams struct {
	SessionID    string          `json:"-"`
	StyleSheetID DOMStyleSheetID `json:"styleSheetId"`
	SelectorText string          `json:"selectorText"`
}

type CSSGetLocationForSelectorResult struct {
	Ranges []CSSSourceRange `json:"ranges"`
}

type CSSTrackComputedStyleUpdatesForNodeParams struct {
	SessionID string     `json:"-"`
	NodeID    *DOMNodeID `json:"nodeId,omitempty"`
}

type CSSTrackComputedStyleUpdatesForNodeResult struct {
}

type CSSTrackComputedStyleUpdatesParams struct {
	SessionID         string                        `json:"-"`
	PropertiesToTrack []CSSCSSComputedStyleProperty `json:"propertiesToTrack"`
}

type CSSTrackComputedStyleUpdatesResult struct {
}

type CSSTakeComputedStyleUpdatesParams struct {
	SessionID string `json:"-"`
}

type CSSTakeComputedStyleUpdatesResult struct {
	// The list of node Ids that have their tracked computed styles updated.
	NodeIds []DOMNodeID `json:"nodeIds"`
}

type CSSSetEffectivePropertyValueForNodeParams struct {
	SessionID string `json:"-"`
	// The element id for which to set property.
	NodeID       DOMNodeID `json:"nodeId"`
	PropertyName string    `json:"propertyName"`
	Value        string    `json:"value"`
}

type CSSSetEffectivePropertyValueForNodeResult struct {
}

type CSSSetPropertyRulePropertyNameParams struct {
	SessionID    string          `json:"-"`
	StyleSheetID DOMStyleSheetID `json:"styleSheetId"`
	Range        CSSSourceRange  `json:"range"`
	PropertyName string          `json:"propertyName"`
}

type CSSSetPropertyRulePropertyNameResult struct {
	// The resulting key text after modification.
	PropertyName CSSValue `json:"propertyName"`
}

type CSSSetKeyframeKeyParams struct {
	SessionID    string          `json:"-"`
	StyleSheetID DOMStyleSheetID `json:"styleSheetId"`
	Range        CSSSourceRange  `json:"range"`
	KeyText      string          `json:"keyText"`
}

type CSSSetKeyframeKeyResult struct {
	// The resulting key text after modification.
	KeyText CSSValue `json:"keyText"`
}

type CSSSetMediaTextParams struct {
	SessionID    string          `json:"-"`
	StyleSheetID DOMStyleSheetID `json:"styleSheetId"`
	Range        CSSSourceRange  `json:"range"`
	Text         string          `json:"text"`
}

type CSSSetMediaTextResult struct {
	// The resulting CSS media rule after modification.
	Media CSSCSSMedia `json:"media"`
}

type CSSSetContainerQueryTextParams struct {
	SessionID    string          `json:"-"`
	StyleSheetID DOMStyleSheetID `json:"styleSheetId"`
	Range        CSSSourceRange  `json:"range"`
	Text         string          `json:"text"`
}

type CSSSetContainerQueryTextResult struct {
	// The resulting CSS container query rule after modification.
	ContainerQuery CSSCSSContainerQuery `json:"containerQuery"`
}

type CSSSetSupportsTextParams struct {
	SessionID    string          `json:"-"`
	StyleSheetID DOMStyleSheetID `json:"styleSheetId"`
	Range        CSSSourceRange  `json:"range"`
	Text         string          `json:"text"`
}

type CSSSetSupportsTextResult struct {
	// The resulting CSS Supports rule after modification.
	Supports CSSCSSSupports `json:"supports"`
}

type CSSSetNavigationTextParams struct {
	SessionID    string          `json:"-"`
	StyleSheetID DOMStyleSheetID `json:"styleSheetId"`
	Range        CSSSourceRange  `json:"range"`
	Text         string          `json:"text"`
}

type CSSSetNavigationTextResult struct {
	// The resulting CSS Navigation rule after modification.
	Navigation CSSCSSNavigation `json:"navigation"`
}

type CSSSetScopeTextParams struct {
	SessionID    string          `json:"-"`
	StyleSheetID DOMStyleSheetID `json:"styleSheetId"`
	Range        CSSSourceRange  `json:"range"`
	Text         string          `json:"text"`
}

type CSSSetScopeTextResult struct {
	// The resulting CSS Scope rule after modification.
	Scope CSSCSSScope `json:"scope"`
}

type CSSSetRuleSelectorParams struct {
	SessionID    string          `json:"-"`
	StyleSheetID DOMStyleSheetID `json:"styleSheetId"`
	Range        CSSSourceRange  `json:"range"`
	Selector     string          `json:"selector"`
}

type CSSSetRuleSelectorResult struct {
	// The resulting selector list after modification.
	SelectorList CSSSelectorList `json:"selectorList"`
}

type CSSSetStyleSheetTextParams struct {
	SessionID    string          `json:"-"`
	StyleSheetID DOMStyleSheetID `json:"styleSheetId"`
	Text         string          `json:"text"`
}

type CSSSetStyleSheetTextResult struct {
	// URL of source map associated with script (if any).
	SourceMapURL *string `json:"sourceMapURL,omitempty"`
}

type CSSSetStyleTextsParams struct {
	SessionID string                    `json:"-"`
	Edits     []CSSStyleDeclarationEdit `json:"edits"`
	// NodeId for the DOM node in whose context custom property declarations for registered properties should be
	NodeForPropertySyntaxValidation *DOMNodeID `json:"nodeForPropertySyntaxValidation,omitempty"`
}

type CSSSetStyleTextsResult struct {
	// The resulting styles after modification.
	Styles []CSSCSSStyle `json:"styles"`
}

type CSSStartRuleUsageTrackingParams struct {
	SessionID string `json:"-"`
}

type CSSStartRuleUsageTrackingResult struct {
}

type CSSStopRuleUsageTrackingParams struct {
	SessionID string `json:"-"`
}

type CSSStopRuleUsageTrackingResult struct {
	RuleUsage []CSSRuleUsage `json:"ruleUsage"`
}

type CSSTakeCoverageDeltaParams struct {
	SessionID string `json:"-"`
}

type CSSTakeCoverageDeltaResult struct {
	Coverage []CSSRuleUsage `json:"coverage"`
	// Monotonically increasing time, in seconds.
	Timestamp float64 `json:"timestamp"`
}

type CSSSetLocalFontsEnabledParams struct {
	SessionID string `json:"-"`
	// Whether rendering of local fonts is enabled.
	Enabled bool `json:"enabled"`
}

type CSSSetLocalFontsEnabledResult struct {
}

type CSSFontsUpdatedEvent struct {
	// The web font that has loaded.
	Font *CSSFontFace `json:"font,omitempty"`
}

type CSSMediaQueryResultChangedEvent struct {
}

type CSSStyleSheetAddedEvent struct {
	// Added stylesheet metainfo.
	Header CSSCSSStyleSheetHeader `json:"header"`
}

type CSSStyleSheetChangedEvent struct {
	StyleSheetID DOMStyleSheetID `json:"styleSheetId"`
}

type CSSStyleSheetRemovedEvent struct {
	// Identifier of the removed stylesheet.
	StyleSheetID DOMStyleSheetID `json:"styleSheetId"`
}

type CSSComputedStyleUpdatedEvent struct {
	// The node id that has updated computed styles.
	NodeID DOMNodeID `json:"nodeId"`
}

type CacheStorageCacheID string

type CacheStorageCachedResponseType string

type CacheStorageDataEntry struct {
	// Request URL.
	RequestURL string `json:"requestURL"`
	// Request method.
	RequestMethod string `json:"requestMethod"`
	// Request headers
	RequestHeaders []CacheStorageHeader `json:"requestHeaders"`
	// Number of seconds since epoch.
	ResponseTime float64 `json:"responseTime"`
	// HTTP response status code.
	ResponseStatus int `json:"responseStatus"`
	// HTTP response status text.
	ResponseStatusText string `json:"responseStatusText"`
	// HTTP response type
	ResponseType CacheStorageCachedResponseType `json:"responseType"`
	// Response headers
	ResponseHeaders []CacheStorageHeader `json:"responseHeaders"`
}

type CacheStorageCache struct {
	// An opaque unique id of the cache.
	CacheID CacheStorageCacheID `json:"cacheId"`
	// Security origin of the cache.
	SecurityOrigin string `json:"securityOrigin"`
	// Storage key of the cache.
	StorageKey string `json:"storageKey"`
	// Storage bucket of the cache.
	StorageBucket *StorageStorageBucket `json:"storageBucket,omitempty"`
	// The name of the cache.
	CacheName string `json:"cacheName"`
}

type CacheStorageHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type CacheStorageCachedResponse struct {
	// Entry content, base64-encoded. (Encoded as a base64 string when passed over JSON)
	Body string `json:"body"`
}

type CacheStorageDeleteCacheParams struct {
	SessionID string `json:"-"`
	// Id of cache for deletion.
	CacheID CacheStorageCacheID `json:"cacheId"`
}

type CacheStorageDeleteCacheResult struct {
}

type CacheStorageDeleteEntryParams struct {
	SessionID string `json:"-"`
	// Id of cache where the entry will be deleted.
	CacheID CacheStorageCacheID `json:"cacheId"`
	// URL spec of the request.
	Request string `json:"request"`
}

type CacheStorageDeleteEntryResult struct {
}

type CacheStorageRequestCacheNamesParams struct {
	SessionID string `json:"-"`
	// At least and at most one of securityOrigin, storageKey, storageBucket must be specified.
	SecurityOrigin *string `json:"securityOrigin,omitempty"`
	// Storage key.
	StorageKey *string `json:"storageKey,omitempty"`
	// Storage bucket. If not specified, it uses the default bucket.
	StorageBucket *StorageStorageBucket `json:"storageBucket,omitempty"`
}

type CacheStorageRequestCacheNamesResult struct {
	// Caches for the security origin.
	Caches []CacheStorageCache `json:"caches"`
}

type CacheStorageRequestCachedResponseParams struct {
	SessionID string `json:"-"`
	// Id of cache that contains the entry.
	CacheID CacheStorageCacheID `json:"cacheId"`
	// URL spec of the request.
	RequestURL string `json:"requestURL"`
	// headers of the request.
	RequestHeaders []CacheStorageHeader `json:"requestHeaders"`
}

type CacheStorageRequestCachedResponseResult struct {
	// Response read from the cache.
	Response CacheStorageCachedResponse `json:"response"`
}

type CacheStorageRequestEntriesParams struct {
	SessionID string `json:"-"`
	// ID of cache to get entries from.
	CacheID CacheStorageCacheID `json:"cacheId"`
	// Number of records to skip.
	SkipCount *int `json:"skipCount,omitempty"`
	// Number of records to fetch.
	PageSize *int `json:"pageSize,omitempty"`
	// If present, only return the entries containing this substring in the path
	PathFilter *string `json:"pathFilter,omitempty"`
}

type CacheStorageRequestEntriesResult struct {
	// Array of object store data entries.
	CacheDataEntries []CacheStorageDataEntry `json:"cacheDataEntries"`
	// Count of returned entries from this storage. If pathFilter is empty, it
	ReturnCount float64 `json:"returnCount"`
}

type CastSink struct {
	Name string `json:"name"`
	ID   string `json:"id"`
	// Text describing the current session. Present only if there is an active
	Session *string `json:"session,omitempty"`
}

type CastEnableParams struct {
	SessionID       string  `json:"-"`
	PresentationURL *string `json:"presentationUrl,omitempty"`
}

type CastEnableResult struct {
}

type CastDisableParams struct {
	SessionID string `json:"-"`
}

type CastDisableResult struct {
}

type CastSetSinkToUseParams struct {
	SessionID string `json:"-"`
	SinkName  string `json:"sinkName"`
}

type CastSetSinkToUseResult struct {
}

type CastStartDesktopMirroringParams struct {
	SessionID string `json:"-"`
	SinkName  string `json:"sinkName"`
}

type CastStartDesktopMirroringResult struct {
}

type CastStartTabMirroringParams struct {
	SessionID string `json:"-"`
	SinkName  string `json:"sinkName"`
}

type CastStartTabMirroringResult struct {
}

type CastStopCastingParams struct {
	SessionID string `json:"-"`
	SinkName  string `json:"sinkName"`
}

type CastStopCastingResult struct {
}

type CastSinksUpdatedEvent struct {
	Sinks []CastSink `json:"sinks"`
}

type CastIssueUpdatedEvent struct {
	IssueMessage string `json:"issueMessage"`
}

type ConsoleConsoleMessage struct {
	// Message source.
	Source string `json:"source"`
	// Message severity.
	Level string `json:"level"`
	// Message text.
	Text string `json:"text"`
	// URL of the message origin.
	URL *string `json:"url,omitempty"`
	// Line number in the resource that generated this message (1-based).
	Line *int `json:"line,omitempty"`
	// Column number in the resource that generated this message (1-based).
	Column *int `json:"column,omitempty"`
}

type ConsoleClearMessagesParams struct {
	SessionID string `json:"-"`
}

type ConsoleClearMessagesResult struct {
}

type ConsoleDisableParams struct {
	SessionID string `json:"-"`
}

type ConsoleDisableResult struct {
}

type ConsoleEnableParams struct {
	SessionID string `json:"-"`
}

type ConsoleEnableResult struct {
}

type ConsoleMessageAddedEvent struct {
	// Console message that has been added.
	Message ConsoleConsoleMessage `json:"message"`
}

type DOMNodeID int

type DOMBackendNodeID int

type DOMStyleSheetID string

type DOMBackendNode struct {
	// `Node`'s nodeType.
	NodeType int `json:"nodeType"`
	// `Node`'s nodeName.
	NodeName      string           `json:"nodeName"`
	BackendNodeID DOMBackendNodeID `json:"backendNodeId"`
}

type DOMPseudoType string

type DOMShadowRootType string

type DOMCompatibilityMode string

type DOMPhysicalAxes string

type DOMLogicalAxes string

type DOMScrollOrientation string

type DOMNode struct {
	// Node identifier that is passed into the rest of the DOM messages as the `nodeId`. Backend
	NodeID DOMNodeID `json:"nodeId"`
	// The id of the parent node if any.
	ParentID *DOMNodeID `json:"parentId,omitempty"`
	// The BackendNodeId for this node.
	BackendNodeID DOMBackendNodeID `json:"backendNodeId"`
	// `Node`'s nodeType.
	NodeType int `json:"nodeType"`
	// `Node`'s nodeName.
	NodeName string `json:"nodeName"`
	// `Node`'s localName.
	LocalName string `json:"localName"`
	// `Node`'s nodeValue.
	NodeValue string `json:"nodeValue"`
	// Child count for `Container` nodes.
	ChildNodeCount *int `json:"childNodeCount,omitempty"`
	// Child nodes of this node when requested with children.
	Children []DOMNode `json:"children,omitempty"`
	// Attributes of the `Element` node in the form of flat array `[name1, value1, name2, value2]`.
	Attributes []string `json:"attributes,omitempty"`
	// Document URL that `Document` or `FrameOwner` node points to.
	DocumentURL *string `json:"documentURL,omitempty"`
	// Base URL that `Document` or `FrameOwner` node uses for URL completion.
	BaseURL *string `json:"baseURL,omitempty"`
	// `DocumentType`'s publicId.
	PublicID *string `json:"publicId,omitempty"`
	// `DocumentType`'s systemId.
	SystemID *string `json:"systemId,omitempty"`
	// `DocumentType`'s internalSubset.
	InternalSubset *string `json:"internalSubset,omitempty"`
	// `Document`'s XML version in case of XML documents.
	XMLVersion *string `json:"xmlVersion,omitempty"`
	// `Attr`'s name.
	Name *string `json:"name,omitempty"`
	// `Attr`'s value.
	Value *string `json:"value,omitempty"`
	// Pseudo element type for this node.
	PseudoType *DOMPseudoType `json:"pseudoType,omitempty"`
	// Pseudo element identifier for this node. Only present if there is a
	PseudoIdentifier *string `json:"pseudoIdentifier,omitempty"`
	// Shadow root type.
	ShadowRootType *DOMShadowRootType `json:"shadowRootType,omitempty"`
	// Frame ID for frame owner elements.
	FrameID *PageFrameID `json:"frameId,omitempty"`
	// Content document for frame owner elements.
	ContentDocument *DOMNode `json:"contentDocument,omitempty"`
	// Shadow root list for given element host.
	ShadowRoots []DOMNode `json:"shadowRoots,omitempty"`
	// Content document fragment for template elements.
	TemplateContent *DOMNode `json:"templateContent,omitempty"`
	// Pseudo elements associated with this node.
	PseudoElements []DOMNode `json:"pseudoElements,omitempty"`
	// Deprecated, as the HTML Imports API has been removed (crbug.com/937746).
	ImportedDocument *DOMNode `json:"importedDocument,omitempty"`
	// Distributed nodes for given insertion point.
	DistributedNodes []DOMBackendNode `json:"distributedNodes,omitempty"`
	// Whether the node is SVG.
	IsSVG                    *bool                 `json:"isSVG,omitempty"`
	CompatibilityMode        *DOMCompatibilityMode `json:"compatibilityMode,omitempty"`
	AssignedSlot             *DOMBackendNode       `json:"assignedSlot,omitempty"`
	IsScrollable             *bool                 `json:"isScrollable,omitempty"`
	AffectedByStartingStyles *bool                 `json:"affectedByStartingStyles,omitempty"`
	AdoptedStyleSheets       []DOMStyleSheetID     `json:"adoptedStyleSheets,omitempty"`
	IsAdRelated              *bool                 `json:"isAdRelated,omitempty"`
}

type DOMDetachedElementInfo struct {
	TreeNode        DOMNode     `json:"treeNode"`
	RetainedNodeIds []DOMNodeID `json:"retainedNodeIds"`
}

type DOMRGBA struct {
	// The red component, in the [0-255] range.
	R int `json:"r"`
	// The green component, in the [0-255] range.
	G int `json:"g"`
	// The blue component, in the [0-255] range.
	B int `json:"b"`
	// The alpha component, in the [0-1] range (default: 1).
	A *float64 `json:"a,omitempty"`
}

type DOMQuad []float64

type DOMBoxModel struct {
	// Content box
	Content DOMQuad `json:"content"`
	// Padding box
	Padding DOMQuad `json:"padding"`
	// Border box
	Border DOMQuad `json:"border"`
	// Margin box
	Margin DOMQuad `json:"margin"`
	// Node width
	Width int `json:"width"`
	// Node height
	Height int `json:"height"`
	// Shape outside coordinates
	ShapeOutside *DOMShapeOutsideInfo `json:"shapeOutside,omitempty"`
}

type DOMShapeOutsideInfo struct {
	// Shape bounds
	Bounds DOMQuad `json:"bounds"`
	// Shape coordinate details
	Shape []any `json:"shape"`
	// Margin shape bounds
	MarginShape []any `json:"marginShape"`
}

type DOMRect struct {
	// X coordinate
	X float64 `json:"x"`
	// Y coordinate
	Y float64 `json:"y"`
	// Rectangle width
	Width float64 `json:"width"`
	// Rectangle height
	Height float64 `json:"height"`
}

type DOMCSSComputedStyleProperty struct {
	// Computed style property name.
	Name string `json:"name"`
	// Computed style property value.
	Value string `json:"value"`
}

type DOMCollectClassNamesFromSubtreeParams struct {
	SessionID string `json:"-"`
	// Id of the node to collect class names.
	NodeID DOMNodeID `json:"nodeId"`
}

type DOMCollectClassNamesFromSubtreeResult struct {
	// Class name list.
	ClassNames []string `json:"classNames"`
}

type DOMCopyToParams struct {
	SessionID string `json:"-"`
	// Id of the node to copy.
	NodeID DOMNodeID `json:"nodeId"`
	// Id of the element to drop the copy into.
	TargetNodeID DOMNodeID `json:"targetNodeId"`
	// Drop the copy before this node (if absent, the copy becomes the last child of
	InsertBeforeNodeID *DOMNodeID `json:"insertBeforeNodeId,omitempty"`
}

type DOMCopyToResult struct {
	// Id of the node clone.
	NodeID DOMNodeID `json:"nodeId"`
}

type DOMDescribeNodeParams struct {
	SessionID string `json:"-"`
	// Identifier of the node.
	NodeID *DOMNodeID `json:"nodeId,omitempty"`
	// Identifier of the backend node.
	BackendNodeID *DOMBackendNodeID `json:"backendNodeId,omitempty"`
	// JavaScript object id of the node wrapper.
	ObjectID *RuntimeRemoteObjectID `json:"objectId,omitempty"`
	// The maximum depth at which children should be retrieved, defaults to 1. Use -1 for the
	Depth *int `json:"depth,omitempty"`
	// Whether or not iframes and shadow roots should be traversed when returning the subtree
	Pierce *bool `json:"pierce,omitempty"`
}

type DOMDescribeNodeResult struct {
	// Node description.
	Node DOMNode `json:"node"`
}

type DOMScrollIntoViewIfNeededParams struct {
	SessionID string `json:"-"`
	// Identifier of the node.
	NodeID *DOMNodeID `json:"nodeId,omitempty"`
	// Identifier of the backend node.
	BackendNodeID *DOMBackendNodeID `json:"backendNodeId,omitempty"`
	// JavaScript object id of the node wrapper.
	ObjectID *RuntimeRemoteObjectID `json:"objectId,omitempty"`
	// The rect to be scrolled into view, relative to the node's border box, in CSS pixels.
	Rect *DOMRect `json:"rect,omitempty"`
}

type DOMScrollIntoViewIfNeededResult struct {
}

type DOMDisableParams struct {
	SessionID string `json:"-"`
}

type DOMDisableResult struct {
}

type DOMDiscardSearchResultsParams struct {
	SessionID string `json:"-"`
	// Unique search session identifier.
	SearchID string `json:"searchId"`
}

type DOMDiscardSearchResultsResult struct {
}

type DOMEnableParams struct {
	SessionID string `json:"-"`
	// Whether to include whitespaces in the children array of returned Nodes.
	IncludeWhitespace *string `json:"includeWhitespace,omitempty"`
}

type DOMEnableResult struct {
}

type DOMFocusParams struct {
	SessionID string `json:"-"`
	// Identifier of the node.
	NodeID *DOMNodeID `json:"nodeId,omitempty"`
	// Identifier of the backend node.
	BackendNodeID *DOMBackendNodeID `json:"backendNodeId,omitempty"`
	// JavaScript object id of the node wrapper.
	ObjectID *RuntimeRemoteObjectID `json:"objectId,omitempty"`
}

type DOMFocusResult struct {
}

type DOMGetAttributesParams struct {
	SessionID string `json:"-"`
	// Id of the node to retrieve attributes for.
	NodeID DOMNodeID `json:"nodeId"`
}

type DOMGetAttributesResult struct {
	// An interleaved array of node attribute names and values.
	Attributes []string `json:"attributes"`
}

type DOMGetBoxModelParams struct {
	SessionID string `json:"-"`
	// Identifier of the node.
	NodeID *DOMNodeID `json:"nodeId,omitempty"`
	// Identifier of the backend node.
	BackendNodeID *DOMBackendNodeID `json:"backendNodeId,omitempty"`
	// JavaScript object id of the node wrapper.
	ObjectID *RuntimeRemoteObjectID `json:"objectId,omitempty"`
}

type DOMGetBoxModelResult struct {
	// Box model for the node.
	Model DOMBoxModel `json:"model"`
}

type DOMGetContentQuadsParams struct {
	SessionID string `json:"-"`
	// Identifier of the node.
	NodeID *DOMNodeID `json:"nodeId,omitempty"`
	// Identifier of the backend node.
	BackendNodeID *DOMBackendNodeID `json:"backendNodeId,omitempty"`
	// JavaScript object id of the node wrapper.
	ObjectID *RuntimeRemoteObjectID `json:"objectId,omitempty"`
}

type DOMGetContentQuadsResult struct {
	// Quads that describe node layout relative to viewport.
	Quads []DOMQuad `json:"quads"`
}

type DOMGetDocumentParams struct {
	SessionID string `json:"-"`
	// The maximum depth at which children should be retrieved, defaults to 1. Use -1 for the
	Depth *int `json:"depth,omitempty"`
	// Whether or not iframes and shadow roots should be traversed when returning the subtree
	Pierce *bool `json:"pierce,omitempty"`
}

type DOMGetDocumentResult struct {
	// Resulting node.
	Root DOMNode `json:"root"`
}

type DOMGetFlattenedDocumentParams struct {
	SessionID string `json:"-"`
	// The maximum depth at which children should be retrieved, defaults to 1. Use -1 for the
	Depth *int `json:"depth,omitempty"`
	// Whether or not iframes and shadow roots should be traversed when returning the subtree
	Pierce *bool `json:"pierce,omitempty"`
}

type DOMGetFlattenedDocumentResult struct {
	// Resulting node.
	Nodes []DOMNode `json:"nodes"`
}

type DOMGetNodesForSubtreeByStyleParams struct {
	SessionID string `json:"-"`
	// Node ID pointing to the root of a subtree.
	NodeID DOMNodeID `json:"nodeId"`
	// The style to filter nodes by (includes nodes if any of properties matches).
	ComputedStyles []DOMCSSComputedStyleProperty `json:"computedStyles"`
	// Whether or not iframes and shadow roots in the same target should be traversed when returning the
	Pierce *bool `json:"pierce,omitempty"`
}

type DOMGetNodesForSubtreeByStyleResult struct {
	// Resulting nodes.
	NodeIds []DOMNodeID `json:"nodeIds"`
}

type DOMGetNodeForLocationParams struct {
	SessionID string `json:"-"`
	// X coordinate.
	X int `json:"x"`
	// Y coordinate.
	Y int `json:"y"`
	// False to skip to the nearest non-UA shadow root ancestor (default: false).
	IncludeUserAgentShadowDOM *bool `json:"includeUserAgentShadowDOM,omitempty"`
	// Whether to ignore pointer-events: none on elements and hit test them.
	IgnorePointerEventsNone *bool `json:"ignorePointerEventsNone,omitempty"`
}

type DOMGetNodeForLocationResult struct {
	// Resulting node.
	BackendNodeID DOMBackendNodeID `json:"backendNodeId"`
	// Frame this node belongs to.
	FrameID PageFrameID `json:"frameId"`
	// Id of the node at given coordinates, only when enabled and requested document.
	NodeID *DOMNodeID `json:"nodeId,omitempty"`
}

type DOMGetOuterHTMLParams struct {
	SessionID string `json:"-"`
	// Identifier of the node.
	NodeID *DOMNodeID `json:"nodeId,omitempty"`
	// Identifier of the backend node.
	BackendNodeID *DOMBackendNodeID `json:"backendNodeId,omitempty"`
	// JavaScript object id of the node wrapper.
	ObjectID *RuntimeRemoteObjectID `json:"objectId,omitempty"`
	// Include all shadow roots. Equals to false if not specified.
	IncludeShadowDOM *bool `json:"includeShadowDOM,omitempty"`
}

type DOMGetOuterHTMLResult struct {
	// Outer HTML markup.
	OuterHTML string `json:"outerHTML"`
}

type DOMGetRelayoutBoundaryParams struct {
	SessionID string `json:"-"`
	// Id of the node.
	NodeID DOMNodeID `json:"nodeId"`
}

type DOMGetRelayoutBoundaryResult struct {
	// Relayout boundary node id for the given node.
	NodeID DOMNodeID `json:"nodeId"`
}

type DOMGetSearchResultsParams struct {
	SessionID string `json:"-"`
	// Unique search session identifier.
	SearchID string `json:"searchId"`
	// Start index of the search result to be returned.
	FromIndex int `json:"fromIndex"`
	// End index of the search result to be returned.
	ToIndex int `json:"toIndex"`
}

type DOMGetSearchResultsResult struct {
	// Ids of the search result nodes.
	NodeIds []DOMNodeID `json:"nodeIds"`
}

type DOMHideHighlightParams struct {
	SessionID string `json:"-"`
}

type DOMHideHighlightResult struct {
}

type DOMHighlightNodeParams struct {
	SessionID string `json:"-"`
}

type DOMHighlightNodeResult struct {
}

type DOMHighlightRectParams struct {
	SessionID string `json:"-"`
}

type DOMHighlightRectResult struct {
}

type DOMMarkUndoableStateParams struct {
	SessionID string `json:"-"`
}

type DOMMarkUndoableStateResult struct {
}

type DOMMoveToParams struct {
	SessionID string `json:"-"`
	// Id of the node to move.
	NodeID DOMNodeID `json:"nodeId"`
	// Id of the element to drop the moved node into.
	TargetNodeID DOMNodeID `json:"targetNodeId"`
	// Drop node before this one (if absent, the moved node becomes the last child of
	InsertBeforeNodeID *DOMNodeID `json:"insertBeforeNodeId,omitempty"`
}

type DOMMoveToResult struct {
	// New id of the moved node.
	NodeID DOMNodeID `json:"nodeId"`
}

type DOMPerformSearchParams struct {
	SessionID string `json:"-"`
	// Plain text or query selector or XPath search query.
	Query string `json:"query"`
	// True to search in user agent shadow DOM.
	IncludeUserAgentShadowDOM *bool `json:"includeUserAgentShadowDOM,omitempty"`
}

type DOMPerformSearchResult struct {
	// Unique search session identifier.
	SearchID string `json:"searchId"`
	// Number of search results.
	ResultCount int `json:"resultCount"`
}

type DOMPushNodeByPathToFrontendParams struct {
	SessionID string `json:"-"`
	// Path to node in the proprietary format.
	Path string `json:"path"`
}

type DOMPushNodeByPathToFrontendResult struct {
	// Id of the node for given path.
	NodeID DOMNodeID `json:"nodeId"`
}

type DOMPushNodesByBackendIdsToFrontendParams struct {
	SessionID string `json:"-"`
	// The array of backend node ids.
	BackendNodeIds []DOMBackendNodeID `json:"backendNodeIds"`
}

type DOMPushNodesByBackendIdsToFrontendResult struct {
	// The array of ids of pushed nodes that correspond to the backend ids specified in
	NodeIds []DOMNodeID `json:"nodeIds"`
}

type DOMQuerySelectorParams struct {
	SessionID string `json:"-"`
	// Id of the node to query upon.
	NodeID DOMNodeID `json:"nodeId"`
	// Selector string.
	Selector string `json:"selector"`
}

type DOMQuerySelectorResult struct {
	// Query selector result.
	NodeID DOMNodeID `json:"nodeId"`
}

type DOMQuerySelectorAllParams struct {
	SessionID string `json:"-"`
	// Id of the node to query upon.
	NodeID DOMNodeID `json:"nodeId"`
	// Selector string.
	Selector string `json:"selector"`
}

type DOMQuerySelectorAllResult struct {
	// Query selector result.
	NodeIds []DOMNodeID `json:"nodeIds"`
}

type DOMGetTopLayerElementsParams struct {
	SessionID string `json:"-"`
}

type DOMGetTopLayerElementsResult struct {
	// NodeIds of top layer elements
	NodeIds []DOMNodeID `json:"nodeIds"`
}

type DOMGetElementByRelationParams struct {
	SessionID string `json:"-"`
	// Id of the node from which to query the relation.
	NodeID DOMNodeID `json:"nodeId"`
	// Type of relation to get.
	Relation string `json:"relation"`
}

type DOMGetElementByRelationResult struct {
	// NodeId of the element matching the queried relation.
	NodeID DOMNodeID `json:"nodeId"`
}

type DOMRedoParams struct {
	SessionID string `json:"-"`
}

type DOMRedoResult struct {
}

type DOMRemoveAttributeParams struct {
	SessionID string `json:"-"`
	// Id of the element to remove attribute from.
	NodeID DOMNodeID `json:"nodeId"`
	// Name of the attribute to remove.
	Name string `json:"name"`
}

type DOMRemoveAttributeResult struct {
}

type DOMRemoveNodeParams struct {
	SessionID string `json:"-"`
	// Id of the node to remove.
	NodeID DOMNodeID `json:"nodeId"`
}

type DOMRemoveNodeResult struct {
}

type DOMRequestChildNodesParams struct {
	SessionID string `json:"-"`
	// Id of the node to get children for.
	NodeID DOMNodeID `json:"nodeId"`
	// The maximum depth at which children should be retrieved, defaults to 1. Use -1 for the
	Depth *int `json:"depth,omitempty"`
	// Whether or not iframes and shadow roots should be traversed when returning the sub-tree
	Pierce *bool `json:"pierce,omitempty"`
}

type DOMRequestChildNodesResult struct {
}

type DOMRequestNodeParams struct {
	SessionID string `json:"-"`
	// JavaScript object id to convert into node.
	ObjectID RuntimeRemoteObjectID `json:"objectId"`
}

type DOMRequestNodeResult struct {
	// Node id for given object.
	NodeID DOMNodeID `json:"nodeId"`
}

type DOMResolveNodeParams struct {
	SessionID string `json:"-"`
	// Id of the node to resolve.
	NodeID *DOMNodeID `json:"nodeId,omitempty"`
	// Backend identifier of the node to resolve.
	BackendNodeID *DOMBackendNodeID `json:"backendNodeId,omitempty"`
	// Symbolic group name that can be used to release multiple objects.
	ObjectGroup *string `json:"objectGroup,omitempty"`
	// Execution context in which to resolve the node.
	ExecutionContextID *RuntimeExecutionContextID `json:"executionContextId,omitempty"`
}

type DOMResolveNodeResult struct {
	// JavaScript object wrapper for given node.
	Object RuntimeRemoteObject `json:"object"`
}

type DOMSetAttributeValueParams struct {
	SessionID string `json:"-"`
	// Id of the element to set attribute for.
	NodeID DOMNodeID `json:"nodeId"`
	// Attribute name.
	Name string `json:"name"`
	// Attribute value.
	Value string `json:"value"`
}

type DOMSetAttributeValueResult struct {
}

type DOMSetAttributesAsTextParams struct {
	SessionID string `json:"-"`
	// Id of the element to set attributes for.
	NodeID DOMNodeID `json:"nodeId"`
	// Text with a number of attributes. Will parse this text using HTML parser.
	Text string `json:"text"`
	// Attribute name to replace with new attributes derived from text in case text parsed
	Name *string `json:"name,omitempty"`
}

type DOMSetAttributesAsTextResult struct {
}

type DOMSetFileInputFilesParams struct {
	SessionID string `json:"-"`
	// Array of file paths to set.
	Files []string `json:"files"`
	// Identifier of the node.
	NodeID *DOMNodeID `json:"nodeId,omitempty"`
	// Identifier of the backend node.
	BackendNodeID *DOMBackendNodeID `json:"backendNodeId,omitempty"`
	// JavaScript object id of the node wrapper.
	ObjectID *RuntimeRemoteObjectID `json:"objectId,omitempty"`
}

type DOMSetFileInputFilesResult struct {
}

type DOMSetNodeStackTracesEnabledParams struct {
	SessionID string `json:"-"`
	// Enable or disable.
	Enable bool `json:"enable"`
}

type DOMSetNodeStackTracesEnabledResult struct {
}

type DOMGetNodeStackTracesParams struct {
	SessionID string `json:"-"`
	// Id of the node to get stack traces for.
	NodeID DOMNodeID `json:"nodeId"`
}

type DOMGetNodeStackTracesResult struct {
	// Creation stack trace, if available.
	Creation *RuntimeStackTrace `json:"creation,omitempty"`
}

type DOMGetFileInfoParams struct {
	SessionID string `json:"-"`
	// JavaScript object id of the node wrapper.
	ObjectID RuntimeRemoteObjectID `json:"objectId"`
}

type DOMGetFileInfoResult struct {
	Path string `json:"path"`
}

type DOMGetDetachedDOMNodesParams struct {
	SessionID string `json:"-"`
}

type DOMGetDetachedDOMNodesResult struct {
	// The list of detached nodes
	DetachedNodes []DOMDetachedElementInfo `json:"detachedNodes"`
}

type DOMSetInspectedNodeParams struct {
	SessionID string `json:"-"`
	// DOM node id to be accessible by means of $x command line API.
	NodeID DOMNodeID `json:"nodeId"`
}

type DOMSetInspectedNodeResult struct {
}

type DOMSetNodeNameParams struct {
	SessionID string `json:"-"`
	// Id of the node to set name for.
	NodeID DOMNodeID `json:"nodeId"`
	// New node's name.
	Name string `json:"name"`
}

type DOMSetNodeNameResult struct {
	// New node's id.
	NodeID DOMNodeID `json:"nodeId"`
}

type DOMSetNodeValueParams struct {
	SessionID string `json:"-"`
	// Id of the node to set value for.
	NodeID DOMNodeID `json:"nodeId"`
	// New node's value.
	Value string `json:"value"`
}

type DOMSetNodeValueResult struct {
}

type DOMSetOuterHTMLParams struct {
	SessionID string `json:"-"`
	// Id of the node to set markup for.
	NodeID DOMNodeID `json:"nodeId"`
	// Outer HTML markup to set.
	OuterHTML string `json:"outerHTML"`
}

type DOMSetOuterHTMLResult struct {
}

type DOMUndoParams struct {
	SessionID string `json:"-"`
}

type DOMUndoResult struct {
}

type DOMGetFrameOwnerParams struct {
	SessionID string      `json:"-"`
	FrameID   PageFrameID `json:"frameId"`
}

type DOMGetFrameOwnerResult struct {
	// Resulting node.
	BackendNodeID DOMBackendNodeID `json:"backendNodeId"`
	// Id of the node at given coordinates, only when enabled and requested document.
	NodeID *DOMNodeID `json:"nodeId,omitempty"`
}

type DOMGetContainerForNodeParams struct {
	SessionID          string           `json:"-"`
	NodeID             DOMNodeID        `json:"nodeId"`
	ContainerName      *string          `json:"containerName,omitempty"`
	PhysicalAxes       *DOMPhysicalAxes `json:"physicalAxes,omitempty"`
	LogicalAxes        *DOMLogicalAxes  `json:"logicalAxes,omitempty"`
	QueriesScrollState *bool            `json:"queriesScrollState,omitempty"`
	QueriesAnchored    *bool            `json:"queriesAnchored,omitempty"`
}

type DOMGetContainerForNodeResult struct {
	// The container node for the given node, or null if not found.
	NodeID *DOMNodeID `json:"nodeId,omitempty"`
}

type DOMGetQueryingDescendantsForContainerParams struct {
	SessionID string `json:"-"`
	// Id of the container node to find querying descendants from.
	NodeID DOMNodeID `json:"nodeId"`
}

type DOMGetQueryingDescendantsForContainerResult struct {
	// Descendant nodes with container queries against the given container.
	NodeIds []DOMNodeID `json:"nodeIds"`
}

type DOMGetAnchorElementParams struct {
	SessionID string `json:"-"`
	// Id of the positioned element from which to find the anchor.
	NodeID DOMNodeID `json:"nodeId"`
	// An optional anchor specifier, as defined in
	AnchorSpecifier *string `json:"anchorSpecifier,omitempty"`
}

type DOMGetAnchorElementResult struct {
	// The anchor element of the given anchor query.
	NodeID DOMNodeID `json:"nodeId"`
}

type DOMForceShowPopoverParams struct {
	SessionID string `json:"-"`
	// Id of the popover HTMLElement
	NodeID DOMNodeID `json:"nodeId"`
	// If true, opens the popover and keeps it open. If false, closes the
	Enable bool `json:"enable"`
}

type DOMForceShowPopoverResult struct {
	// List of popovers that were closed in order to respect popover stacking order.
	NodeIds []DOMNodeID `json:"nodeIds"`
}

type DOMAttributeModifiedEvent struct {
	// Id of the node that has changed.
	NodeID DOMNodeID `json:"nodeId"`
	// Attribute name.
	Name string `json:"name"`
	// Attribute value.
	Value string `json:"value"`
}

type DOMAdoptedStyleSheetsModifiedEvent struct {
	// Id of the node that has changed.
	NodeID DOMNodeID `json:"nodeId"`
	// New adoptedStyleSheets array.
	AdoptedStyleSheets []DOMStyleSheetID `json:"adoptedStyleSheets"`
}

type DOMAttributeRemovedEvent struct {
	// Id of the node that has changed.
	NodeID DOMNodeID `json:"nodeId"`
	// A ttribute name.
	Name string `json:"name"`
}

type DOMCharacterDataModifiedEvent struct {
	// Id of the node that has changed.
	NodeID DOMNodeID `json:"nodeId"`
	// New text value.
	CharacterData string `json:"characterData"`
}

type DOMChildNodeCountUpdatedEvent struct {
	// Id of the node that has changed.
	NodeID DOMNodeID `json:"nodeId"`
	// New node count.
	ChildNodeCount int `json:"childNodeCount"`
}

type DOMChildNodeInsertedEvent struct {
	// Id of the node that has changed.
	ParentNodeID DOMNodeID `json:"parentNodeId"`
	// Id of the previous sibling.
	PreviousNodeID DOMNodeID `json:"previousNodeId"`
	// Inserted node data.
	Node DOMNode `json:"node"`
}

type DOMChildNodeRemovedEvent struct {
	// Parent id.
	ParentNodeID DOMNodeID `json:"parentNodeId"`
	// Id of the node that has been removed.
	NodeID DOMNodeID `json:"nodeId"`
}

type DOMDistributedNodesUpdatedEvent struct {
	// Insertion point where distributed nodes were updated.
	InsertionPointID DOMNodeID `json:"insertionPointId"`
	// Distributed nodes for given insertion point.
	DistributedNodes []DOMBackendNode `json:"distributedNodes"`
}

type DOMDocumentUpdatedEvent struct {
}

type DOMInlineStyleInvalidatedEvent struct {
	// Ids of the nodes for which the inline styles have been invalidated.
	NodeIds []DOMNodeID `json:"nodeIds"`
}

type DOMPseudoElementAddedEvent struct {
	// Pseudo element's parent element id.
	ParentID DOMNodeID `json:"parentId"`
	// The added pseudo element.
	PseudoElement DOMNode `json:"pseudoElement"`
}

type DOMTopLayerElementsUpdatedEvent struct {
}

type DOMScrollableFlagUpdatedEvent struct {
	// The id of the node.
	NodeID DOMNodeID `json:"nodeId"`
	// If the node is scrollable.
	IsScrollable bool `json:"isScrollable"`
}

type DOMAdRelatedStateUpdatedEvent struct {
	// The id of the node.
	NodeID DOMNodeID `json:"nodeId"`
	// If the node is ad related.
	IsAdRelated bool `json:"isAdRelated"`
}

type DOMAffectedByStartingStylesFlagUpdatedEvent struct {
	// The id of the node.
	NodeID DOMNodeID `json:"nodeId"`
	// If the node has starting styles.
	AffectedByStartingStyles bool `json:"affectedByStartingStyles"`
}

type DOMPseudoElementRemovedEvent struct {
	// Pseudo element's parent element id.
	ParentID DOMNodeID `json:"parentId"`
	// The removed pseudo element id.
	PseudoElementID DOMNodeID `json:"pseudoElementId"`
}

type DOMSetChildNodesEvent struct {
	// Parent node id to populate with children.
	ParentID DOMNodeID `json:"parentId"`
	// Child nodes array.
	Nodes []DOMNode `json:"nodes"`
}

type DOMShadowRootPoppedEvent struct {
	// Host element id.
	HostID DOMNodeID `json:"hostId"`
	// Shadow root id.
	RootID DOMNodeID `json:"rootId"`
}

type DOMShadowRootPushedEvent struct {
	// Host element id.
	HostID DOMNodeID `json:"hostId"`
	// Shadow root.
	Root DOMNode `json:"root"`
}

type DOMDebuggerDOMBreakpointType string

type DOMDebuggerCSPViolationType string

type DOMDebuggerEventListener struct {
	// `EventListener`'s type.
	Type string `json:"type"`
	// `EventListener`'s useCapture.
	UseCapture bool `json:"useCapture"`
	// `EventListener`'s passive flag.
	Passive bool `json:"passive"`
	// `EventListener`'s once flag.
	Once bool `json:"once"`
	// Script id of the handler code.
	ScriptID RuntimeScriptID `json:"scriptId"`
	// Line number in the script (0-based).
	LineNumber int `json:"lineNumber"`
	// Column number in the script (0-based).
	ColumnNumber int `json:"columnNumber"`
	// Event handler function value.
	Handler *RuntimeRemoteObject `json:"handler,omitempty"`
	// Event original handler function value.
	OriginalHandler *RuntimeRemoteObject `json:"originalHandler,omitempty"`
	// Node the listener is added to (if any).
	BackendNodeID *DOMBackendNodeID `json:"backendNodeId,omitempty"`
}

type DOMDebuggerGetEventListenersParams struct {
	SessionID string `json:"-"`
	// Identifier of the object to return listeners for.
	ObjectID RuntimeRemoteObjectID `json:"objectId"`
	// The maximum depth at which Node children should be retrieved, defaults to 1. Use -1 for the
	Depth *int `json:"depth,omitempty"`
	// Whether or not iframes and shadow roots should be traversed when returning the subtree
	Pierce *bool `json:"pierce,omitempty"`
}

type DOMDebuggerGetEventListenersResult struct {
	// Array of relevant listeners.
	Listeners []DOMDebuggerEventListener `json:"listeners"`
}

type DOMDebuggerRemoveDOMBreakpointParams struct {
	SessionID string `json:"-"`
	// Identifier of the node to remove breakpoint from.
	NodeID DOMNodeID `json:"nodeId"`
	// Type of the breakpoint to remove.
	Type DOMDebuggerDOMBreakpointType `json:"type"`
}

type DOMDebuggerRemoveDOMBreakpointResult struct {
}

type DOMDebuggerRemoveEventListenerBreakpointParams struct {
	SessionID string `json:"-"`
	// Event name.
	EventName string `json:"eventName"`
	// EventTarget interface name.
	TargetName *string `json:"targetName,omitempty"`
}

type DOMDebuggerRemoveEventListenerBreakpointResult struct {
}

type DOMDebuggerRemoveInstrumentationBreakpointParams struct {
	SessionID string `json:"-"`
	// Instrumentation name to stop on.
	EventName string `json:"eventName"`
}

type DOMDebuggerRemoveInstrumentationBreakpointResult struct {
}

type DOMDebuggerRemoveXHRBreakpointParams struct {
	SessionID string `json:"-"`
	// Resource URL substring.
	URL string `json:"url"`
}

type DOMDebuggerRemoveXHRBreakpointResult struct {
}

type DOMDebuggerSetBreakOnCSPViolationParams struct {
	SessionID string `json:"-"`
	// CSP Violations to stop upon.
	ViolationTypes []DOMDebuggerCSPViolationType `json:"violationTypes"`
}

type DOMDebuggerSetBreakOnCSPViolationResult struct {
}

type DOMDebuggerSetDOMBreakpointParams struct {
	SessionID string `json:"-"`
	// Identifier of the node to set breakpoint on.
	NodeID DOMNodeID `json:"nodeId"`
	// Type of the operation to stop upon.
	Type DOMDebuggerDOMBreakpointType `json:"type"`
}

type DOMDebuggerSetDOMBreakpointResult struct {
}

type DOMDebuggerSetEventListenerBreakpointParams struct {
	SessionID string `json:"-"`
	// DOM Event name to stop on (any DOM event will do).
	EventName string `json:"eventName"`
	// EventTarget interface name to stop on. If equal to `"*"` or not provided, will stop on any
	TargetName *string `json:"targetName,omitempty"`
}

type DOMDebuggerSetEventListenerBreakpointResult struct {
}

type DOMDebuggerSetInstrumentationBreakpointParams struct {
	SessionID string `json:"-"`
	// Instrumentation name to stop on.
	EventName string `json:"eventName"`
}

type DOMDebuggerSetInstrumentationBreakpointResult struct {
}

type DOMDebuggerSetXHRBreakpointParams struct {
	SessionID string `json:"-"`
	// Resource URL substring. All XHRs having this substring in the URL will get stopped upon.
	URL string `json:"url"`
}

type DOMDebuggerSetXHRBreakpointResult struct {
}

type DOMSnapshotDOMNode struct {
	// `Node`'s nodeType.
	NodeType int `json:"nodeType"`
	// `Node`'s nodeName.
	NodeName string `json:"nodeName"`
	// `Node`'s nodeValue.
	NodeValue string `json:"nodeValue"`
	// Only set for textarea elements, contains the text value.
	TextValue *string `json:"textValue,omitempty"`
	// Only set for input elements, contains the input's associated text value.
	InputValue *string `json:"inputValue,omitempty"`
	// Only set for radio and checkbox input elements, indicates if the element has been checked
	InputChecked *bool `json:"inputChecked,omitempty"`
	// Only set for option elements, indicates if the element has been selected
	OptionSelected *bool `json:"optionSelected,omitempty"`
	// `Node`'s id, corresponds to DOM.Node.backendNodeId.
	BackendNodeID DOMBackendNodeID `json:"backendNodeId"`
	// The indexes of the node's child nodes in the `domNodes` array returned by `getSnapshot`, if
	ChildNodeIndexes []int `json:"childNodeIndexes,omitempty"`
	// Attributes of an `Element` node.
	Attributes []DOMSnapshotNameValue `json:"attributes,omitempty"`
	// Indexes of pseudo elements associated with this node in the `domNodes` array returned by
	PseudoElementIndexes []int `json:"pseudoElementIndexes,omitempty"`
	// The index of the node's related layout tree node in the `layoutTreeNodes` array returned by
	LayoutNodeIndex *int `json:"layoutNodeIndex,omitempty"`
	// Document URL that `Document` or `FrameOwner` node points to.
	DocumentURL *string `json:"documentURL,omitempty"`
	// Base URL that `Document` or `FrameOwner` node uses for URL completion.
	BaseURL *string `json:"baseURL,omitempty"`
	// Only set for documents, contains the document's content language.
	ContentLanguage *string `json:"contentLanguage,omitempty"`
	// Only set for documents, contains the document's character set encoding.
	DocumentEncoding *string `json:"documentEncoding,omitempty"`
	// `DocumentType` node's publicId.
	PublicID *string `json:"publicId,omitempty"`
	// `DocumentType` node's systemId.
	SystemID *string `json:"systemId,omitempty"`
	// Frame ID for frame owner elements and also for the document node.
	FrameID *PageFrameID `json:"frameId,omitempty"`
	// The index of a frame owner element's content document in the `domNodes` array returned by
	ContentDocumentIndex *int `json:"contentDocumentIndex,omitempty"`
	// Type of a pseudo element node.
	PseudoType *DOMPseudoType `json:"pseudoType,omitempty"`
	// Shadow root type.
	ShadowRootType *DOMShadowRootType `json:"shadowRootType,omitempty"`
	// Whether this DOM node responds to mouse clicks. This includes nodes that have had click
	IsClickable *bool `json:"isClickable,omitempty"`
	// Details of the node's event listeners, if any.
	EventListeners []DOMDebuggerEventListener `json:"eventListeners,omitempty"`
	// The selected url for nodes with a srcset attribute.
	CurrentSourceURL *string `json:"currentSourceURL,omitempty"`
	// The url of the script (if any) that generates this node.
	OriginURL *string `json:"originURL,omitempty"`
	// Scroll offsets, set when this node is a Document.
	ScrollOffsetX *float64 `json:"scrollOffsetX,omitempty"`
	ScrollOffsetY *float64 `json:"scrollOffsetY,omitempty"`
}

type DOMSnapshotInlineTextBox struct {
	// The bounding box in document coordinates. Note that scroll offset of the document is ignored.
	BoundingBox DOMRect `json:"boundingBox"`
	// The starting index in characters, for this post layout textbox substring. Characters that
	StartCharacterIndex int `json:"startCharacterIndex"`
	// The number of characters in this post layout textbox substring. Characters that would be
	NumCharacters int `json:"numCharacters"`
}

type DOMSnapshotLayoutTreeNode struct {
	// The index of the related DOM node in the `domNodes` array returned by `getSnapshot`.
	DOMNodeIndex int `json:"domNodeIndex"`
	// The bounding box in document coordinates. Note that scroll offset of the document is ignored.
	BoundingBox DOMRect `json:"boundingBox"`
	// Contents of the LayoutText, if any.
	LayoutText *string `json:"layoutText,omitempty"`
	// The post-layout inline text nodes, if any.
	InlineTextNodes []DOMSnapshotInlineTextBox `json:"inlineTextNodes,omitempty"`
	// Index into the `computedStyles` array returned by `getSnapshot`.
	StyleIndex *int `json:"styleIndex,omitempty"`
	// Global paint order index, which is determined by the stacking order of the nodes. Nodes
	PaintOrder *int `json:"paintOrder,omitempty"`
	// Set to true to indicate the element begins a new stacking context.
	IsStackingContext *bool `json:"isStackingContext,omitempty"`
}

type DOMSnapshotComputedStyle struct {
	// Name/value pairs of computed style properties.
	Properties []DOMSnapshotNameValue `json:"properties"`
}

type DOMSnapshotNameValue struct {
	// Attribute/property name.
	Name string `json:"name"`
	// Attribute/property value.
	Value string `json:"value"`
}

type DOMSnapshotStringIndex int

type DOMSnapshotArrayOfStrings []DOMSnapshotStringIndex

type DOMSnapshotRareStringData struct {
	Index []int                    `json:"index"`
	Value []DOMSnapshotStringIndex `json:"value"`
}

type DOMSnapshotRareBooleanData struct {
	Index []int `json:"index"`
}

type DOMSnapshotRareIntegerData struct {
	Index []int `json:"index"`
	Value []int `json:"value"`
}

type DOMSnapshotRectangle []float64

type DOMSnapshotDocumentSnapshot struct {
	// Document URL that `Document` or `FrameOwner` node points to.
	DocumentURL DOMSnapshotStringIndex `json:"documentURL"`
	// Document title.
	Title DOMSnapshotStringIndex `json:"title"`
	// Base URL that `Document` or `FrameOwner` node uses for URL completion.
	BaseURL DOMSnapshotStringIndex `json:"baseURL"`
	// Contains the document's content language.
	ContentLanguage DOMSnapshotStringIndex `json:"contentLanguage"`
	// Contains the document's character set encoding.
	EncodingName DOMSnapshotStringIndex `json:"encodingName"`
	// `DocumentType` node's publicId.
	PublicID DOMSnapshotStringIndex `json:"publicId"`
	// `DocumentType` node's systemId.
	SystemID DOMSnapshotStringIndex `json:"systemId"`
	// Frame ID for frame owner elements and also for the document node.
	FrameID DOMSnapshotStringIndex `json:"frameId"`
	// A table with dom nodes.
	Nodes DOMSnapshotNodeTreeSnapshot `json:"nodes"`
	// The nodes in the layout tree.
	Layout DOMSnapshotLayoutTreeSnapshot `json:"layout"`
	// The post-layout inline text nodes.
	TextBoxes DOMSnapshotTextBoxSnapshot `json:"textBoxes"`
	// Horizontal scroll offset.
	ScrollOffsetX *float64 `json:"scrollOffsetX,omitempty"`
	// Vertical scroll offset.
	ScrollOffsetY *float64 `json:"scrollOffsetY,omitempty"`
	// Document content width.
	ContentWidth *float64 `json:"contentWidth,omitempty"`
	// Document content height.
	ContentHeight *float64 `json:"contentHeight,omitempty"`
}

type DOMSnapshotNodeTreeSnapshot struct {
	// Parent node index.
	ParentIndex []int `json:"parentIndex,omitempty"`
	// `Node`'s nodeType.
	NodeType []int `json:"nodeType,omitempty"`
	// Type of the shadow root the `Node` is in. String values are equal to the `ShadowRootType` enum.
	ShadowRootType *DOMSnapshotRareStringData `json:"shadowRootType,omitempty"`
	// `Node`'s nodeName.
	NodeName []DOMSnapshotStringIndex `json:"nodeName,omitempty"`
	// `Node`'s nodeValue.
	NodeValue []DOMSnapshotStringIndex `json:"nodeValue,omitempty"`
	// `Node`'s id, corresponds to DOM.Node.backendNodeId.
	BackendNodeID []DOMBackendNodeID `json:"backendNodeId,omitempty"`
	// Attributes of an `Element` node. Flatten name, value pairs.
	Attributes []DOMSnapshotArrayOfStrings `json:"attributes,omitempty"`
	// Only set for textarea elements, contains the text value.
	TextValue *DOMSnapshotRareStringData `json:"textValue,omitempty"`
	// Only set for input elements, contains the input's associated text value.
	InputValue *DOMSnapshotRareStringData `json:"inputValue,omitempty"`
	// Only set for radio and checkbox input elements, indicates if the element has been checked
	InputChecked *DOMSnapshotRareBooleanData `json:"inputChecked,omitempty"`
	// Only set for option elements, indicates if the element has been selected
	OptionSelected *DOMSnapshotRareBooleanData `json:"optionSelected,omitempty"`
	// The index of the document in the list of the snapshot documents.
	ContentDocumentIndex *DOMSnapshotRareIntegerData `json:"contentDocumentIndex,omitempty"`
	// Type of a pseudo element node.
	PseudoType *DOMSnapshotRareStringData `json:"pseudoType,omitempty"`
	// Pseudo element identifier for this node. Only present if there is a
	PseudoIdentifier *DOMSnapshotRareStringData `json:"pseudoIdentifier,omitempty"`
	// Whether this DOM node responds to mouse clicks. This includes nodes that have had click
	IsClickable *DOMSnapshotRareBooleanData `json:"isClickable,omitempty"`
	// The selected url for nodes with a srcset attribute.
	CurrentSourceURL *DOMSnapshotRareStringData `json:"currentSourceURL,omitempty"`
	// The url of the script (if any) that generates this node.
	OriginURL *DOMSnapshotRareStringData `json:"originURL,omitempty"`
}

type DOMSnapshotLayoutTreeSnapshot struct {
	// Index of the corresponding node in the `NodeTreeSnapshot` array returned by `captureSnapshot`.
	NodeIndex []int `json:"nodeIndex"`
	// Array of indexes specifying computed style strings, filtered according to the `computedStyles` parameter passed to `captureSnapshot`.
	Styles []DOMSnapshotArrayOfStrings `json:"styles"`
	// The absolute position bounding box.
	Bounds []DOMSnapshotRectangle `json:"bounds"`
	// Contents of the LayoutText, if any.
	Text []DOMSnapshotStringIndex `json:"text"`
	// Stacking context information.
	StackingContexts DOMSnapshotRareBooleanData `json:"stackingContexts"`
	// Global paint order index, which is determined by the stacking order of the nodes. Nodes
	PaintOrders []int `json:"paintOrders,omitempty"`
	// The offset rect of nodes. Only available when includeDOMRects is set to true
	OffsetRects []DOMSnapshotRectangle `json:"offsetRects,omitempty"`
	// The scroll rect of nodes. Only available when includeDOMRects is set to true
	ScrollRects []DOMSnapshotRectangle `json:"scrollRects,omitempty"`
	// The client rect of nodes. Only available when includeDOMRects is set to true
	ClientRects []DOMSnapshotRectangle `json:"clientRects,omitempty"`
	// The list of background colors that are blended with colors of overlapping elements.
	BlendedBackgroundColors []DOMSnapshotStringIndex `json:"blendedBackgroundColors,omitempty"`
	// The list of computed text opacities.
	TextColorOpacities []float64 `json:"textColorOpacities,omitempty"`
}

type DOMSnapshotTextBoxSnapshot struct {
	// Index of the layout tree node that owns this box collection.
	LayoutIndex []int `json:"layoutIndex"`
	// The absolute position bounding box.
	Bounds []DOMSnapshotRectangle `json:"bounds"`
	// The starting index in characters, for this post layout textbox substring. Characters that
	Start []int `json:"start"`
	// The number of characters in this post layout textbox substring. Characters that would be
	Length []int `json:"length"`
}

type DOMSnapshotDisableParams struct {
	SessionID string `json:"-"`
}

type DOMSnapshotDisableResult struct {
}

type DOMSnapshotEnableParams struct {
	SessionID string `json:"-"`
}

type DOMSnapshotEnableResult struct {
}

type DOMSnapshotGetSnapshotParams struct {
	SessionID string `json:"-"`
	// Whitelist of computed styles to return.
	ComputedStyleWhitelist []string `json:"computedStyleWhitelist"`
	// Whether or not to retrieve details of DOM listeners (default false).
	IncludeEventListeners *bool `json:"includeEventListeners,omitempty"`
	// Whether to determine and include the paint order index of LayoutTreeNodes (default false).
	IncludePaintOrder *bool `json:"includePaintOrder,omitempty"`
	// Whether to include UA shadow tree in the snapshot (default false).
	IncludeUserAgentShadowTree *bool `json:"includeUserAgentShadowTree,omitempty"`
}

type DOMSnapshotGetSnapshotResult struct {
	// The nodes in the DOM tree. The DOMNode at index 0 corresponds to the root document.
	DOMNodes []DOMSnapshotDOMNode `json:"domNodes"`
	// The nodes in the layout tree.
	LayoutTreeNodes []DOMSnapshotLayoutTreeNode `json:"layoutTreeNodes"`
	// Whitelisted ComputedStyle properties for each node in the layout tree.
	ComputedStyles []DOMSnapshotComputedStyle `json:"computedStyles"`
}

type DOMSnapshotCaptureSnapshotParams struct {
	SessionID string `json:"-"`
	// Whitelist of computed styles to return.
	ComputedStyles []string `json:"computedStyles"`
	// Whether to include layout object paint orders into the snapshot.
	IncludePaintOrder *bool `json:"includePaintOrder,omitempty"`
	// Whether to include DOM rectangles (offsetRects, clientRects, scrollRects) into the snapshot
	IncludeDOMRects *bool `json:"includeDOMRects,omitempty"`
	// Whether to include blended background colors in the snapshot (default: false).
	IncludeBlendedBackgroundColors *bool `json:"includeBlendedBackgroundColors,omitempty"`
	// Whether to include text color opacity in the snapshot (default: false).
	IncludeTextColorOpacities *bool `json:"includeTextColorOpacities,omitempty"`
}

type DOMSnapshotCaptureSnapshotResult struct {
	// The nodes in the DOM tree. The DOMNode at index 0 corresponds to the root document.
	Documents []DOMSnapshotDocumentSnapshot `json:"documents"`
	// Shared string table that all string properties refer to with indexes.
	Strings []string `json:"strings"`
}

type DOMStorageSerializedStorageKey string

type DOMStorageStorageID struct {
	// Security origin for the storage.
	SecurityOrigin *string `json:"securityOrigin,omitempty"`
	// Represents a key by which DOM Storage keys its CachedStorageAreas
	StorageKey *DOMStorageSerializedStorageKey `json:"storageKey,omitempty"`
	// Whether the storage is local storage (not session storage).
	IsLocalStorage bool `json:"isLocalStorage"`
}

type DOMStorageItem []string

type DOMStorageClearParams struct {
	SessionID string              `json:"-"`
	StorageID DOMStorageStorageID `json:"storageId"`
}

type DOMStorageClearResult struct {
}

type DOMStorageDisableParams struct {
	SessionID string `json:"-"`
}

type DOMStorageDisableResult struct {
}

type DOMStorageEnableParams struct {
	SessionID string `json:"-"`
}

type DOMStorageEnableResult struct {
}

type DOMStorageGetDOMStorageItemsParams struct {
	SessionID string              `json:"-"`
	StorageID DOMStorageStorageID `json:"storageId"`
}

type DOMStorageGetDOMStorageItemsResult struct {
	Entries []DOMStorageItem `json:"entries"`
}

type DOMStorageRemoveDOMStorageItemParams struct {
	SessionID string              `json:"-"`
	StorageID DOMStorageStorageID `json:"storageId"`
	Key       string              `json:"key"`
}

type DOMStorageRemoveDOMStorageItemResult struct {
}

type DOMStorageSetDOMStorageItemParams struct {
	SessionID string              `json:"-"`
	StorageID DOMStorageStorageID `json:"storageId"`
	Key       string              `json:"key"`
	Value     string              `json:"value"`
}

type DOMStorageSetDOMStorageItemResult struct {
}

type DOMStorageDOMStorageItemAddedEvent struct {
	StorageID DOMStorageStorageID `json:"storageId"`
	Key       string              `json:"key"`
	NewValue  string              `json:"newValue"`
}

type DOMStorageDOMStorageItemRemovedEvent struct {
	StorageID DOMStorageStorageID `json:"storageId"`
	Key       string              `json:"key"`
}

type DOMStorageDOMStorageItemUpdatedEvent struct {
	StorageID DOMStorageStorageID `json:"storageId"`
	Key       string              `json:"key"`
	OldValue  string              `json:"oldValue"`
	NewValue  string              `json:"newValue"`
}

type DOMStorageDOMStorageItemsClearedEvent struct {
	StorageID DOMStorageStorageID `json:"storageId"`
}

type DebuggerBreakpointID string

type DebuggerCallFrameID string

type DebuggerLocation struct {
	// Script identifier as reported in the `Debugger.scriptParsed`.
	ScriptID RuntimeScriptID `json:"scriptId"`
	// Line number in the script (0-based).
	LineNumber int `json:"lineNumber"`
	// Column number in the script (0-based).
	ColumnNumber *int `json:"columnNumber,omitempty"`
}

type DebuggerScriptPosition struct {
	LineNumber   int `json:"lineNumber"`
	ColumnNumber int `json:"columnNumber"`
}

type DebuggerLocationRange struct {
	ScriptID RuntimeScriptID        `json:"scriptId"`
	Start    DebuggerScriptPosition `json:"start"`
	End      DebuggerScriptPosition `json:"end"`
}

type DebuggerCallFrame struct {
	// Call frame identifier. This identifier is only valid while the virtual machine is paused.
	CallFrameID DebuggerCallFrameID `json:"callFrameId"`
	// Name of the JavaScript function called on this call frame.
	FunctionName string `json:"functionName"`
	// Location in the source code.
	FunctionLocation *DebuggerLocation `json:"functionLocation,omitempty"`
	// Location in the source code.
	Location DebuggerLocation `json:"location"`
	// JavaScript script name or url.
	URL string `json:"url"`
	// Scope chain for this call frame.
	ScopeChain []DebuggerScope `json:"scopeChain"`
	// `this` object for this call frame.
	This RuntimeRemoteObject `json:"this"`
	// The value being returned, if the function is at return point.
	ReturnValue *RuntimeRemoteObject `json:"returnValue,omitempty"`
	// Valid only while the VM is paused and indicates whether this frame
	CanBeRestarted *bool `json:"canBeRestarted,omitempty"`
}

type DebuggerScope struct {
	// Scope type.
	Type string `json:"type"`
	// Object representing the scope. For `global` and `with` scopes it represents the actual
	Object RuntimeRemoteObject `json:"object"`
	Name   *string             `json:"name,omitempty"`
	// Location in the source code where scope starts
	StartLocation *DebuggerLocation `json:"startLocation,omitempty"`
	// Location in the source code where scope ends
	EndLocation *DebuggerLocation `json:"endLocation,omitempty"`
}

type DebuggerSearchMatch struct {
	// Line number in resource content.
	LineNumber float64 `json:"lineNumber"`
	// Line with match content.
	LineContent string `json:"lineContent"`
}

type DebuggerBreakLocation struct {
	// Script identifier as reported in the `Debugger.scriptParsed`.
	ScriptID RuntimeScriptID `json:"scriptId"`
	// Line number in the script (0-based).
	LineNumber int `json:"lineNumber"`
	// Column number in the script (0-based).
	ColumnNumber *int    `json:"columnNumber,omitempty"`
	Type         *string `json:"type,omitempty"`
}

type DebuggerWasmDisassemblyChunk struct {
	// The next chunk of disassembled lines.
	Lines []string `json:"lines"`
	// The bytecode offsets describing the start of each line.
	BytecodeOffsets []int `json:"bytecodeOffsets"`
}

type DebuggerScriptLanguage string

type DebuggerDebugSymbols struct {
	// Type of the debug symbols.
	Type string `json:"type"`
	// URL of the external symbol source.
	ExternalURL *string `json:"externalURL,omitempty"`
}

type DebuggerResolvedBreakpoint struct {
	// Breakpoint unique identifier.
	BreakpointID DebuggerBreakpointID `json:"breakpointId"`
	// Actual breakpoint location.
	Location DebuggerLocation `json:"location"`
}

type DebuggerContinueToLocationParams struct {
	SessionID string `json:"-"`
	// Location to continue to.
	Location         DebuggerLocation `json:"location"`
	TargetCallFrames *string          `json:"targetCallFrames,omitempty"`
}

type DebuggerContinueToLocationResult struct {
}

type DebuggerDisableParams struct {
	SessionID string `json:"-"`
}

type DebuggerDisableResult struct {
}

type DebuggerEnableParams struct {
	SessionID string `json:"-"`
	// The maximum size in bytes of collected scripts (not referenced by other heap objects)
	MaxScriptsCacheSize *float64 `json:"maxScriptsCacheSize,omitempty"`
}

type DebuggerEnableResult struct {
	// Unique identifier of the debugger.
	DebuggerID RuntimeUniqueDebuggerID `json:"debuggerId"`
}

type DebuggerEvaluateOnCallFrameParams struct {
	SessionID string `json:"-"`
	// Call frame identifier to evaluate on.
	CallFrameID DebuggerCallFrameID `json:"callFrameId"`
	// Expression to evaluate.
	Expression string `json:"expression"`
	// String object group name to put result into (allows rapid releasing resulting object handles
	ObjectGroup *string `json:"objectGroup,omitempty"`
	// Specifies whether command line API should be available to the evaluated expression, defaults
	IncludeCommandLineAPI *bool `json:"includeCommandLineAPI,omitempty"`
	// In silent mode exceptions thrown during evaluation are not reported and do not pause
	Silent *bool `json:"silent,omitempty"`
	// Whether the result is expected to be a JSON object that should be sent by value.
	ReturnByValue *bool `json:"returnByValue,omitempty"`
	// Whether preview should be generated for the result.
	GeneratePreview *bool `json:"generatePreview,omitempty"`
	// Whether to throw an exception if side effect cannot be ruled out during evaluation.
	ThrowOnSideEffect *bool `json:"throwOnSideEffect,omitempty"`
	// Terminate execution after timing out (number of milliseconds).
	Timeout *RuntimeTimeDelta `json:"timeout,omitempty"`
}

type DebuggerEvaluateOnCallFrameResult struct {
	// Object wrapper for the evaluation result.
	Result RuntimeRemoteObject `json:"result"`
	// Exception details.
	ExceptionDetails *RuntimeExceptionDetails `json:"exceptionDetails,omitempty"`
}

type DebuggerGetPossibleBreakpointsParams struct {
	SessionID string `json:"-"`
	// Start of range to search possible breakpoint locations in.
	Start DebuggerLocation `json:"start"`
	// End of range to search possible breakpoint locations in (excluding). When not specified, end
	End *DebuggerLocation `json:"end,omitempty"`
	// Only consider locations which are in the same (non-nested) function as start.
	RestrictToFunction *bool `json:"restrictToFunction,omitempty"`
}

type DebuggerGetPossibleBreakpointsResult struct {
	// List of the possible breakpoint locations.
	Locations []DebuggerBreakLocation `json:"locations"`
}

type DebuggerGetScriptSourceParams struct {
	SessionID string `json:"-"`
	// Id of the script to get source for.
	ScriptID RuntimeScriptID `json:"scriptId"`
}

type DebuggerGetScriptSourceResult struct {
	// Script source (empty in case of Wasm bytecode).
	ScriptSource string `json:"scriptSource"`
	// Wasm bytecode. (Encoded as a base64 string when passed over JSON)
	Bytecode *string `json:"bytecode,omitempty"`
}

type DebuggerDisassembleWasmModuleParams struct {
	SessionID string `json:"-"`
	// Id of the script to disassemble
	ScriptID RuntimeScriptID `json:"scriptId"`
}

type DebuggerDisassembleWasmModuleResult struct {
	// For large modules, return a stream from which additional chunks of
	StreamID *string `json:"streamId,omitempty"`
	// The total number of lines in the disassembly text.
	TotalNumberOfLines int `json:"totalNumberOfLines"`
	// The offsets of all function bodies, in the format [start1, end1,
	FunctionBodyOffsets []int `json:"functionBodyOffsets"`
	// The first chunk of disassembly.
	Chunk DebuggerWasmDisassemblyChunk `json:"chunk"`
}

type DebuggerNextWasmDisassemblyChunkParams struct {
	SessionID string `json:"-"`
	StreamID  string `json:"streamId"`
}

type DebuggerNextWasmDisassemblyChunkResult struct {
	// The next chunk of disassembly.
	Chunk DebuggerWasmDisassemblyChunk `json:"chunk"`
}

type DebuggerGetWasmBytecodeParams struct {
	SessionID string `json:"-"`
	// Id of the Wasm script to get source for.
	ScriptID RuntimeScriptID `json:"scriptId"`
}

type DebuggerGetWasmBytecodeResult struct {
	// Script source. (Encoded as a base64 string when passed over JSON)
	Bytecode string `json:"bytecode"`
}

type DebuggerGetStackTraceParams struct {
	SessionID    string              `json:"-"`
	StackTraceID RuntimeStackTraceID `json:"stackTraceId"`
}

type DebuggerGetStackTraceResult struct {
	StackTrace RuntimeStackTrace `json:"stackTrace"`
}

type DebuggerPauseParams struct {
	SessionID string `json:"-"`
}

type DebuggerPauseResult struct {
}

type DebuggerPauseOnAsyncCallParams struct {
	SessionID string `json:"-"`
	// Debugger will pause when async call with given stack trace is started.
	ParentStackTraceID RuntimeStackTraceID `json:"parentStackTraceId"`
}

type DebuggerPauseOnAsyncCallResult struct {
}

type DebuggerRemoveBreakpointParams struct {
	SessionID    string               `json:"-"`
	BreakpointID DebuggerBreakpointID `json:"breakpointId"`
}

type DebuggerRemoveBreakpointResult struct {
}

type DebuggerRestartFrameParams struct {
	SessionID string `json:"-"`
	// Call frame identifier to evaluate on.
	CallFrameID DebuggerCallFrameID `json:"callFrameId"`
	// The `mode` parameter must be present and set to 'StepInto', otherwise
	Mode *string `json:"mode,omitempty"`
}

type DebuggerRestartFrameResult struct {
	// New stack trace.
	CallFrames []DebuggerCallFrame `json:"callFrames"`
	// Async stack trace, if any.
	AsyncStackTrace *RuntimeStackTrace `json:"asyncStackTrace,omitempty"`
	// Async stack trace, if any.
	AsyncStackTraceID *RuntimeStackTraceID `json:"asyncStackTraceId,omitempty"`
}

type DebuggerResumeParams struct {
	SessionID string `json:"-"`
	// Set to true to terminate execution upon resuming execution. In contrast
	TerminateOnResume *bool `json:"terminateOnResume,omitempty"`
}

type DebuggerResumeResult struct {
}

type DebuggerSearchInContentParams struct {
	SessionID string `json:"-"`
	// Id of the script to search in.
	ScriptID RuntimeScriptID `json:"scriptId"`
	// String to search for.
	Query string `json:"query"`
	// If true, search is case sensitive.
	CaseSensitive *bool `json:"caseSensitive,omitempty"`
	// If true, treats string parameter as regex.
	IsRegex *bool `json:"isRegex,omitempty"`
}

type DebuggerSearchInContentResult struct {
	// List of search matches.
	Result []DebuggerSearchMatch `json:"result"`
}

type DebuggerSetAsyncCallStackDepthParams struct {
	SessionID string `json:"-"`
	// Maximum depth of async call stacks. Setting to `0` will effectively disable collecting async
	MaxDepth int `json:"maxDepth"`
}

type DebuggerSetAsyncCallStackDepthResult struct {
}

type DebuggerSetBlackboxExecutionContextsParams struct {
	SessionID string `json:"-"`
	// Array of execution context unique ids for the debugger to ignore.
	UniqueIds []string `json:"uniqueIds"`
}

type DebuggerSetBlackboxExecutionContextsResult struct {
}

type DebuggerSetBlackboxPatternsParams struct {
	SessionID string `json:"-"`
	// Array of regexps that will be used to check script url for blackbox state.
	Patterns []string `json:"patterns"`
	// If true, also ignore scripts with no source url.
	SkipAnonymous *bool `json:"skipAnonymous,omitempty"`
}

type DebuggerSetBlackboxPatternsResult struct {
}

type DebuggerSetBlackboxedRangesParams struct {
	SessionID string `json:"-"`
	// Id of the script.
	ScriptID  RuntimeScriptID          `json:"scriptId"`
	Positions []DebuggerScriptPosition `json:"positions"`
}

type DebuggerSetBlackboxedRangesResult struct {
}

type DebuggerSetBreakpointParams struct {
	SessionID string `json:"-"`
	// Location to set breakpoint in.
	Location DebuggerLocation `json:"location"`
	// Expression to use as a breakpoint condition. When specified, debugger will only stop on the
	Condition *string `json:"condition,omitempty"`
}

type DebuggerSetBreakpointResult struct {
	// Id of the created breakpoint for further reference.
	BreakpointID DebuggerBreakpointID `json:"breakpointId"`
	// Location this breakpoint resolved into.
	ActualLocation DebuggerLocation `json:"actualLocation"`
}

type DebuggerSetInstrumentationBreakpointParams struct {
	SessionID string `json:"-"`
	// Instrumentation name.
	Instrumentation string `json:"instrumentation"`
}

type DebuggerSetInstrumentationBreakpointResult struct {
	// Id of the created breakpoint for further reference.
	BreakpointID DebuggerBreakpointID `json:"breakpointId"`
}

type DebuggerSetBreakpointByURLParams struct {
	SessionID string `json:"-"`
	// Line number to set breakpoint at.
	LineNumber int `json:"lineNumber"`
	// URL of the resources to set breakpoint on.
	URL *string `json:"url,omitempty"`
	// Regex pattern for the URLs of the resources to set breakpoints on. Either `url` or
	URLRegex *string `json:"urlRegex,omitempty"`
	// Script hash of the resources to set breakpoint on.
	ScriptHash *string `json:"scriptHash,omitempty"`
	// Offset in the line to set breakpoint at.
	ColumnNumber *int `json:"columnNumber,omitempty"`
	// Expression to use as a breakpoint condition. When specified, debugger will only stop on the
	Condition *string `json:"condition,omitempty"`
}

type DebuggerSetBreakpointByURLResult struct {
	// Id of the created breakpoint for further reference.
	BreakpointID DebuggerBreakpointID `json:"breakpointId"`
	// List of the locations this breakpoint resolved into upon addition.
	Locations []DebuggerLocation `json:"locations"`
}

type DebuggerSetBreakpointOnFunctionCallParams struct {
	SessionID string `json:"-"`
	// Function object id.
	ObjectID RuntimeRemoteObjectID `json:"objectId"`
	// Expression to use as a breakpoint condition. When specified, debugger will
	Condition *string `json:"condition,omitempty"`
}

type DebuggerSetBreakpointOnFunctionCallResult struct {
	// Id of the created breakpoint for further reference.
	BreakpointID DebuggerBreakpointID `json:"breakpointId"`
}

type DebuggerSetBreakpointsActiveParams struct {
	SessionID string `json:"-"`
	// New value for breakpoints active state.
	Active bool `json:"active"`
}

type DebuggerSetBreakpointsActiveResult struct {
}

type DebuggerSetPauseOnExceptionsParams struct {
	SessionID string `json:"-"`
	// Pause on exceptions mode.
	State string `json:"state"`
}

type DebuggerSetPauseOnExceptionsResult struct {
}

type DebuggerSetReturnValueParams struct {
	SessionID string `json:"-"`
	// New return value.
	NewValue RuntimeCallArgument `json:"newValue"`
}

type DebuggerSetReturnValueResult struct {
}

type DebuggerSetScriptSourceParams struct {
	SessionID string `json:"-"`
	// Id of the script to edit.
	ScriptID RuntimeScriptID `json:"scriptId"`
	// New content of the script.
	ScriptSource string `json:"scriptSource"`
	// If true the change will not actually be applied. Dry run may be used to get result
	DryRun *bool `json:"dryRun,omitempty"`
	// If true, then `scriptSource` is allowed to change the function on top of the stack
	AllowTopFrameEditing *bool `json:"allowTopFrameEditing,omitempty"`
}

type DebuggerSetScriptSourceResult struct {
	// New stack trace in case editing has happened while VM was stopped.
	CallFrames []DebuggerCallFrame `json:"callFrames,omitempty"`
	// Whether current call stack  was modified after applying the changes.
	StackChanged *bool `json:"stackChanged,omitempty"`
	// Async stack trace, if any.
	AsyncStackTrace *RuntimeStackTrace `json:"asyncStackTrace,omitempty"`
	// Async stack trace, if any.
	AsyncStackTraceID *RuntimeStackTraceID `json:"asyncStackTraceId,omitempty"`
	// Whether the operation was successful or not. Only `Ok` denotes a
	Status string `json:"status"`
	// Exception details if any. Only present when `status` is `CompileError`.
	ExceptionDetails *RuntimeExceptionDetails `json:"exceptionDetails,omitempty"`
}

type DebuggerSetSkipAllPausesParams struct {
	SessionID string `json:"-"`
	// New value for skip pauses state.
	Skip bool `json:"skip"`
}

type DebuggerSetSkipAllPausesResult struct {
}

type DebuggerSetVariableValueParams struct {
	SessionID string `json:"-"`
	// 0-based number of scope as was listed in scope chain. Only 'local', 'closure' and 'catch'
	ScopeNumber int `json:"scopeNumber"`
	// Variable name.
	VariableName string `json:"variableName"`
	// New variable value.
	NewValue RuntimeCallArgument `json:"newValue"`
	// Id of callframe that holds variable.
	CallFrameID DebuggerCallFrameID `json:"callFrameId"`
}

type DebuggerSetVariableValueResult struct {
}

type DebuggerStepIntoParams struct {
	SessionID string `json:"-"`
	// Debugger will pause on the execution of the first async task which was scheduled
	BreakOnAsyncCall *bool `json:"breakOnAsyncCall,omitempty"`
	// The skipList specifies location ranges that should be skipped on step into.
	SkipList []DebuggerLocationRange `json:"skipList,omitempty"`
}

type DebuggerStepIntoResult struct {
}

type DebuggerStepOutParams struct {
	SessionID string `json:"-"`
}

type DebuggerStepOutResult struct {
}

type DebuggerStepOverParams struct {
	SessionID string `json:"-"`
	// The skipList specifies location ranges that should be skipped on step over.
	SkipList []DebuggerLocationRange `json:"skipList,omitempty"`
}

type DebuggerStepOverResult struct {
}

type DebuggerBreakpointResolvedEvent struct {
	// Breakpoint unique identifier.
	BreakpointID DebuggerBreakpointID `json:"breakpointId"`
	// Actual breakpoint location.
	Location DebuggerLocation `json:"location"`
}

type DebuggerPausedEvent struct {
	// Call stack the virtual machine stopped on.
	CallFrames []DebuggerCallFrame `json:"callFrames"`
	// Pause reason.
	Reason string `json:"reason"`
	// Object containing break-specific auxiliary properties.
	Data map[string]any `json:"data,omitempty"`
	// Hit breakpoints IDs
	HitBreakpoints []string `json:"hitBreakpoints,omitempty"`
	// Async stack trace, if any.
	AsyncStackTrace *RuntimeStackTrace `json:"asyncStackTrace,omitempty"`
	// Async stack trace, if any.
	AsyncStackTraceID *RuntimeStackTraceID `json:"asyncStackTraceId,omitempty"`
	// Never present, will be removed.
	AsyncCallStackTraceID *RuntimeStackTraceID `json:"asyncCallStackTraceId,omitempty"`
}

type DebuggerResumedEvent struct {
}

type DebuggerScriptFailedToParseEvent struct {
	// Identifier of the script parsed.
	ScriptID RuntimeScriptID `json:"scriptId"`
	// URL or name of the script parsed (if any).
	URL string `json:"url"`
	// Line offset of the script within the resource with given URL (for script tags).
	StartLine int `json:"startLine"`
	// Column offset of the script within the resource with given URL.
	StartColumn int `json:"startColumn"`
	// Last line of the script.
	EndLine int `json:"endLine"`
	// Length of the last line of the script.
	EndColumn int `json:"endColumn"`
	// Specifies script creation context.
	ExecutionContextID RuntimeExecutionContextID `json:"executionContextId"`
	// Content hash of the script, SHA-256.
	Hash string `json:"hash"`
	// For Wasm modules, the content of the `build_id` custom section. For JavaScript the `debugId` magic comment.
	BuildID string `json:"buildId"`
	// Embedder-specific auxiliary data likely matching {isDefault: boolean, type: 'default'|'isolated'|'worker', frameId: string}
	ExecutionContextAuxData map[string]any `json:"executionContextAuxData,omitempty"`
	// URL of source map associated with script (if any).
	SourceMapURL *string `json:"sourceMapURL,omitempty"`
	// True, if this script has sourceURL.
	HasSourceURL *bool `json:"hasSourceURL,omitempty"`
	// True, if this script is ES6 module.
	IsModule *bool `json:"isModule,omitempty"`
	// This script length.
	Length *int `json:"length,omitempty"`
	// JavaScript top stack frame of where the script parsed event was triggered if available.
	StackTrace *RuntimeStackTrace `json:"stackTrace,omitempty"`
	// If the scriptLanguage is WebAssembly, the code section offset in the module.
	CodeOffset *int `json:"codeOffset,omitempty"`
	// The language of the script.
	ScriptLanguage *DebuggerScriptLanguage `json:"scriptLanguage,omitempty"`
	// The name the embedder supplied for this script.
	EmbedderName *string `json:"embedderName,omitempty"`
}

type DebuggerScriptParsedEvent struct {
	// Identifier of the script parsed.
	ScriptID RuntimeScriptID `json:"scriptId"`
	// URL or name of the script parsed (if any).
	URL string `json:"url"`
	// Line offset of the script within the resource with given URL (for script tags).
	StartLine int `json:"startLine"`
	// Column offset of the script within the resource with given URL.
	StartColumn int `json:"startColumn"`
	// Last line of the script.
	EndLine int `json:"endLine"`
	// Length of the last line of the script.
	EndColumn int `json:"endColumn"`
	// Specifies script creation context.
	ExecutionContextID RuntimeExecutionContextID `json:"executionContextId"`
	// Content hash of the script, SHA-256.
	Hash string `json:"hash"`
	// For Wasm modules, the content of the `build_id` custom section. For JavaScript the `debugId` magic comment.
	BuildID string `json:"buildId"`
	// Embedder-specific auxiliary data likely matching {isDefault: boolean, type: 'default'|'isolated'|'worker', frameId: string}
	ExecutionContextAuxData map[string]any `json:"executionContextAuxData,omitempty"`
	// True, if this script is generated as a result of the live edit operation.
	IsLiveEdit *bool `json:"isLiveEdit,omitempty"`
	// URL of source map associated with script (if any).
	SourceMapURL *string `json:"sourceMapURL,omitempty"`
	// True, if this script has sourceURL.
	HasSourceURL *bool `json:"hasSourceURL,omitempty"`
	// True, if this script is ES6 module.
	IsModule *bool `json:"isModule,omitempty"`
	// This script length.
	Length *int `json:"length,omitempty"`
	// JavaScript top stack frame of where the script parsed event was triggered if available.
	StackTrace *RuntimeStackTrace `json:"stackTrace,omitempty"`
	// If the scriptLanguage is WebAssembly, the code section offset in the module.
	CodeOffset *int `json:"codeOffset,omitempty"`
	// The language of the script.
	ScriptLanguage *DebuggerScriptLanguage `json:"scriptLanguage,omitempty"`
	// If the scriptLanguage is WebAssembly, the source of debug symbols for the module.
	DebugSymbols []DebuggerDebugSymbols `json:"debugSymbols,omitempty"`
	// The name the embedder supplied for this script.
	EmbedderName *string `json:"embedderName,omitempty"`
	// The list of set breakpoints in this script if calls to `setBreakpointByUrl`
	ResolvedBreakpoints []DebuggerResolvedBreakpoint `json:"resolvedBreakpoints,omitempty"`
}

type DeviceAccessRequestID string

type DeviceAccessDeviceID string

type DeviceAccessPromptDevice struct {
	ID DeviceAccessDeviceID `json:"id"`
	// Display name as it appears in a device request user prompt.
	Name string `json:"name"`
}

type DeviceAccessEnableParams struct {
	SessionID string `json:"-"`
}

type DeviceAccessEnableResult struct {
}

type DeviceAccessDisableParams struct {
	SessionID string `json:"-"`
}

type DeviceAccessDisableResult struct {
}

type DeviceAccessSelectPromptParams struct {
	SessionID string                `json:"-"`
	ID        DeviceAccessRequestID `json:"id"`
	DeviceID  DeviceAccessDeviceID  `json:"deviceId"`
}

type DeviceAccessSelectPromptResult struct {
}

type DeviceAccessCancelPromptParams struct {
	SessionID string                `json:"-"`
	ID        DeviceAccessRequestID `json:"id"`
}

type DeviceAccessCancelPromptResult struct {
}

type DeviceAccessDeviceRequestPromptedEvent struct {
	ID      DeviceAccessRequestID      `json:"id"`
	Devices []DeviceAccessPromptDevice `json:"devices"`
}

type DeviceOrientationClearDeviceOrientationOverrideParams struct {
	SessionID string `json:"-"`
}

type DeviceOrientationClearDeviceOrientationOverrideResult struct {
}

type DeviceOrientationSetDeviceOrientationOverrideParams struct {
	SessionID string `json:"-"`
	// Mock alpha
	Alpha float64 `json:"alpha"`
	// Mock beta
	Beta float64 `json:"beta"`
	// Mock gamma
	Gamma float64 `json:"gamma"`
}

type DeviceOrientationSetDeviceOrientationOverrideResult struct {
}

type EmulationSafeAreaInsets struct {
	// Overrides safe-area-inset-top.
	Top *int `json:"top,omitempty"`
	// Overrides safe-area-max-inset-top.
	TopMax *int `json:"topMax,omitempty"`
	// Overrides safe-area-inset-left.
	Left *int `json:"left,omitempty"`
	// Overrides safe-area-max-inset-left.
	LeftMax *int `json:"leftMax,omitempty"`
	// Overrides safe-area-inset-bottom.
	Bottom *int `json:"bottom,omitempty"`
	// Overrides safe-area-max-inset-bottom.
	BottomMax *int `json:"bottomMax,omitempty"`
	// Overrides safe-area-inset-right.
	Right *int `json:"right,omitempty"`
	// Overrides safe-area-max-inset-right.
	RightMax *int `json:"rightMax,omitempty"`
}

type EmulationScreenOrientation struct {
	// Orientation type.
	Type string `json:"type"`
	// Orientation angle.
	Angle int `json:"angle"`
}

type EmulationDisplayFeature struct {
	// Orientation of a display feature in relation to screen
	Orientation string `json:"orientation"`
	// The offset from the screen origin in either the x (for vertical
	Offset int `json:"offset"`
	// A display feature may mask content such that it is not physically
	MaskLength int `json:"maskLength"`
}

type EmulationDevicePosture struct {
	// Current posture of the device
	Type string `json:"type"`
}

type EmulationMediaFeature struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type EmulationVirtualTimePolicy string

type EmulationUserAgentBrandVersion struct {
	Brand   string `json:"brand"`
	Version string `json:"version"`
}

type EmulationUserAgentMetadata struct {
	// Brands appearing in Sec-CH-UA.
	Brands []EmulationUserAgentBrandVersion `json:"brands,omitempty"`
	// Brands appearing in Sec-CH-UA-Full-Version-List.
	FullVersionList []EmulationUserAgentBrandVersion `json:"fullVersionList,omitempty"`
	FullVersion     *string                          `json:"fullVersion,omitempty"`
	Platform        string                           `json:"platform"`
	PlatformVersion string                           `json:"platformVersion"`
	Architecture    string                           `json:"architecture"`
	Model           string                           `json:"model"`
	Mobile          bool                             `json:"mobile"`
	Bitness         *string                          `json:"bitness,omitempty"`
	Wow64           *bool                            `json:"wow64,omitempty"`
	// Used to specify User Agent form-factor values.
	FormFactors []string `json:"formFactors,omitempty"`
}

type EmulationSensorType string

type EmulationSensorMetadata struct {
	Available        *bool    `json:"available,omitempty"`
	MinimumFrequency *float64 `json:"minimumFrequency,omitempty"`
	MaximumFrequency *float64 `json:"maximumFrequency,omitempty"`
}

type EmulationSensorReadingSingle struct {
	Value float64 `json:"value"`
}

type EmulationSensorReadingXYZ struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

type EmulationSensorReadingQuaternion struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
	W float64 `json:"w"`
}

type EmulationSensorReading struct {
	Single     *EmulationSensorReadingSingle     `json:"single,omitempty"`
	Xyz        *EmulationSensorReadingXYZ        `json:"xyz,omitempty"`
	Quaternion *EmulationSensorReadingQuaternion `json:"quaternion,omitempty"`
}

type EmulationPressureSource string

type EmulationPressureState string

type EmulationPressureMetadata struct {
	Available *bool `json:"available,omitempty"`
}

type EmulationWorkAreaInsets struct {
	// Work area top inset in pixels. Default is 0;
	Top *int `json:"top,omitempty"`
	// Work area left inset in pixels. Default is 0;
	Left *int `json:"left,omitempty"`
	// Work area bottom inset in pixels. Default is 0;
	Bottom *int `json:"bottom,omitempty"`
	// Work area right inset in pixels. Default is 0;
	Right *int `json:"right,omitempty"`
}

type EmulationScreenID string

type EmulationScreenInfo struct {
	// Offset of the left edge of the screen.
	Left int `json:"left"`
	// Offset of the top edge of the screen.
	Top int `json:"top"`
	// Width of the screen.
	Width int `json:"width"`
	// Height of the screen.
	Height int `json:"height"`
	// Offset of the left edge of the available screen area.
	AvailLeft int `json:"availLeft"`
	// Offset of the top edge of the available screen area.
	AvailTop int `json:"availTop"`
	// Width of the available screen area.
	AvailWidth int `json:"availWidth"`
	// Height of the available screen area.
	AvailHeight int `json:"availHeight"`
	// Specifies the screen's device pixel ratio.
	DevicePixelRatio float64 `json:"devicePixelRatio"`
	// Specifies the screen's orientation.
	Orientation EmulationScreenOrientation `json:"orientation"`
	// Specifies the screen's color depth in bits.
	ColorDepth int `json:"colorDepth"`
	// Indicates whether the device has multiple screens.
	IsExtended bool `json:"isExtended"`
	// Indicates whether the screen is internal to the device or external, attached to the device.
	IsInternal bool `json:"isInternal"`
	// Indicates whether the screen is set as the the operating system primary screen.
	IsPrimary bool `json:"isPrimary"`
	// Specifies the descriptive label for the screen.
	Label string `json:"label"`
	// Specifies the unique identifier of the screen.
	ID EmulationScreenID `json:"id"`
}

type EmulationDisabledImageType string

type EmulationCanEmulateParams struct {
	SessionID string `json:"-"`
}

type EmulationCanEmulateResult struct {
	// True if emulation is supported.
	Result bool `json:"result"`
}

type EmulationClearDeviceMetricsOverrideParams struct {
	SessionID string `json:"-"`
}

type EmulationClearDeviceMetricsOverrideResult struct {
}

type EmulationClearGeolocationOverrideParams struct {
	SessionID string `json:"-"`
}

type EmulationClearGeolocationOverrideResult struct {
}

type EmulationResetPageScaleFactorParams struct {
	SessionID string `json:"-"`
}

type EmulationResetPageScaleFactorResult struct {
}

type EmulationSetFocusEmulationEnabledParams struct {
	SessionID string `json:"-"`
	// Whether to enable to disable focus emulation.
	Enabled bool `json:"enabled"`
}

type EmulationSetFocusEmulationEnabledResult struct {
}

type EmulationSetAutoDarkModeOverrideParams struct {
	SessionID string `json:"-"`
	// Whether to enable or disable automatic dark mode.
	Enabled *bool `json:"enabled,omitempty"`
}

type EmulationSetAutoDarkModeOverrideResult struct {
}

type EmulationSetCPUThrottlingRateParams struct {
	SessionID string `json:"-"`
	// Throttling rate as a slowdown factor (1 is no throttle, 2 is 2x slowdown, etc).
	Rate float64 `json:"rate"`
}

type EmulationSetCPUThrottlingRateResult struct {
}

type EmulationSetDefaultBackgroundColorOverrideParams struct {
	SessionID string `json:"-"`
	// RGBA of the default background color. If not specified, any existing override will be
	Color *DOMRGBA `json:"color,omitempty"`
}

type EmulationSetDefaultBackgroundColorOverrideResult struct {
}

type EmulationSetSafeAreaInsetsOverrideParams struct {
	SessionID string                  `json:"-"`
	Insets    EmulationSafeAreaInsets `json:"insets"`
}

type EmulationSetSafeAreaInsetsOverrideResult struct {
}

type EmulationSetDeviceMetricsOverrideParams struct {
	SessionID string `json:"-"`
	// Overriding width value in pixels (minimum 0, maximum 10000000). 0 disables the override.
	Width int `json:"width"`
	// Overriding height value in pixels (minimum 0, maximum 10000000). 0 disables the override.
	Height int `json:"height"`
	// Overriding device scale factor value. 0 disables the override.
	DeviceScaleFactor float64 `json:"deviceScaleFactor"`
	// Whether to emulate mobile device. This includes viewport meta tag, overlay scrollbars, text
	Mobile bool `json:"mobile"`
	// Scale to apply to resulting view image.
	Scale *float64 `json:"scale,omitempty"`
	// Overriding screen width value in pixels (minimum 0, maximum 10000000).
	ScreenWidth *int `json:"screenWidth,omitempty"`
	// Overriding screen height value in pixels (minimum 0, maximum 10000000).
	ScreenHeight *int `json:"screenHeight,omitempty"`
	// Overriding view X position on screen in pixels (minimum 0, maximum 10000000).
	PositionX *int `json:"positionX,omitempty"`
	// Overriding view Y position on screen in pixels (minimum 0, maximum 10000000).
	PositionY *int `json:"positionY,omitempty"`
	// Do not set visible view size, rely upon explicit setVisibleSize call.
	DontSetVisibleSize *bool `json:"dontSetVisibleSize,omitempty"`
	// Screen orientation override.
	ScreenOrientation *EmulationScreenOrientation `json:"screenOrientation,omitempty"`
	// If set, the visible area of the page will be overridden to this viewport. This viewport
	Viewport *PageViewport `json:"viewport,omitempty"`
	// If set, the display feature of a multi-segment screen. If not set, multi-segment support
	DisplayFeature *EmulationDisplayFeature `json:"displayFeature,omitempty"`
	// If set, the posture of a foldable device. If not set the posture is set
	DevicePosture *EmulationDevicePosture `json:"devicePosture,omitempty"`
	// Scrollbar type. Default: `default`.
	ScrollbarType *string `json:"scrollbarType,omitempty"`
	// If set to true, enables screen orientation lock emulation, which
	ScreenOrientationLockEmulation *bool `json:"screenOrientationLockEmulation,omitempty"`
}

type EmulationSetDeviceMetricsOverrideResult struct {
}

type EmulationSetDevicePostureOverrideParams struct {
	SessionID string                 `json:"-"`
	Posture   EmulationDevicePosture `json:"posture"`
}

type EmulationSetDevicePostureOverrideResult struct {
}

type EmulationClearDevicePostureOverrideParams struct {
	SessionID string `json:"-"`
}

type EmulationClearDevicePostureOverrideResult struct {
}

type EmulationSetDisplayFeaturesOverrideParams struct {
	SessionID string                    `json:"-"`
	Features  []EmulationDisplayFeature `json:"features"`
}

type EmulationSetDisplayFeaturesOverrideResult struct {
}

type EmulationClearDisplayFeaturesOverrideParams struct {
	SessionID string `json:"-"`
}

type EmulationClearDisplayFeaturesOverrideResult struct {
}

type EmulationSetScrollbarsHiddenParams struct {
	SessionID string `json:"-"`
	// Whether scrollbars should be always hidden.
	Hidden bool `json:"hidden"`
}

type EmulationSetScrollbarsHiddenResult struct {
}

type EmulationSetDocumentCookieDisabledParams struct {
	SessionID string `json:"-"`
	// Whether document.coookie API should be disabled.
	Disabled bool `json:"disabled"`
}

type EmulationSetDocumentCookieDisabledResult struct {
}

type EmulationSetEmitTouchEventsForMouseParams struct {
	SessionID string `json:"-"`
	// Whether touch emulation based on mouse input should be enabled.
	Enabled bool `json:"enabled"`
	// Touch/gesture events configuration. Default: current platform.
	Configuration *string `json:"configuration,omitempty"`
}

type EmulationSetEmitTouchEventsForMouseResult struct {
}

type EmulationSetEmulatedMediaParams struct {
	SessionID string `json:"-"`
	// Media type to emulate. Empty string disables the override.
	Media *string `json:"media,omitempty"`
	// Media features to emulate.
	Features []EmulationMediaFeature `json:"features,omitempty"`
}

type EmulationSetEmulatedMediaResult struct {
}

type EmulationSetEmulatedVisionDeficiencyParams struct {
	SessionID string `json:"-"`
	// Vision deficiency to emulate. Order: best-effort emulations come first, followed by any
	Type string `json:"type"`
}

type EmulationSetEmulatedVisionDeficiencyResult struct {
}

type EmulationSetEmulatedOSTextScaleParams struct {
	SessionID string   `json:"-"`
	Scale     *float64 `json:"scale,omitempty"`
}

type EmulationSetEmulatedOSTextScaleResult struct {
}

type EmulationSetGeolocationOverrideParams struct {
	SessionID string `json:"-"`
	// Mock latitude
	Latitude *float64 `json:"latitude,omitempty"`
	// Mock longitude
	Longitude *float64 `json:"longitude,omitempty"`
	// Mock accuracy
	Accuracy *float64 `json:"accuracy,omitempty"`
	// Mock altitude
	Altitude *float64 `json:"altitude,omitempty"`
	// Mock altitudeAccuracy
	AltitudeAccuracy *float64 `json:"altitudeAccuracy,omitempty"`
	// Mock heading
	Heading *float64 `json:"heading,omitempty"`
	// Mock speed
	Speed *float64 `json:"speed,omitempty"`
}

type EmulationSetGeolocationOverrideResult struct {
}

type EmulationGetOverriddenSensorInformationParams struct {
	SessionID string              `json:"-"`
	Type      EmulationSensorType `json:"type"`
}

type EmulationGetOverriddenSensorInformationResult struct {
	RequestedSamplingFrequency float64 `json:"requestedSamplingFrequency"`
}

type EmulationSetSensorOverrideEnabledParams struct {
	SessionID string                   `json:"-"`
	Enabled   bool                     `json:"enabled"`
	Type      EmulationSensorType      `json:"type"`
	Metadata  *EmulationSensorMetadata `json:"metadata,omitempty"`
}

type EmulationSetSensorOverrideEnabledResult struct {
}

type EmulationSetSensorOverrideReadingsParams struct {
	SessionID string                 `json:"-"`
	Type      EmulationSensorType    `json:"type"`
	Reading   EmulationSensorReading `json:"reading"`
}

type EmulationSetSensorOverrideReadingsResult struct {
}

type EmulationSetPressureSourceOverrideEnabledParams struct {
	SessionID string                     `json:"-"`
	Enabled   bool                       `json:"enabled"`
	Source    EmulationPressureSource    `json:"source"`
	Metadata  *EmulationPressureMetadata `json:"metadata,omitempty"`
}

type EmulationSetPressureSourceOverrideEnabledResult struct {
}

type EmulationSetPressureStateOverrideParams struct {
	SessionID string                  `json:"-"`
	Source    EmulationPressureSource `json:"source"`
	State     EmulationPressureState  `json:"state"`
}

type EmulationSetPressureStateOverrideResult struct {
}

type EmulationSetPressureDataOverrideParams struct {
	SessionID               string                  `json:"-"`
	Source                  EmulationPressureSource `json:"source"`
	State                   EmulationPressureState  `json:"state"`
	OwnContributionEstimate *float64                `json:"ownContributionEstimate,omitempty"`
}

type EmulationSetPressureDataOverrideResult struct {
}

type EmulationSetIdleOverrideParams struct {
	SessionID string `json:"-"`
	// Mock isUserActive
	IsUserActive bool `json:"isUserActive"`
	// Mock isScreenUnlocked
	IsScreenUnlocked bool `json:"isScreenUnlocked"`
}

type EmulationSetIdleOverrideResult struct {
}

type EmulationClearIdleOverrideParams struct {
	SessionID string `json:"-"`
}

type EmulationClearIdleOverrideResult struct {
}

type EmulationSetNavigatorOverridesParams struct {
	SessionID string `json:"-"`
	// The platform navigator.platform should return.
	Platform string `json:"platform"`
}

type EmulationSetNavigatorOverridesResult struct {
}

type EmulationSetPageScaleFactorParams struct {
	SessionID string `json:"-"`
	// Page scale factor.
	PageScaleFactor float64 `json:"pageScaleFactor"`
}

type EmulationSetPageScaleFactorResult struct {
}

type EmulationSetScriptExecutionDisabledParams struct {
	SessionID string `json:"-"`
	// Whether script execution should be disabled in the page.
	Value bool `json:"value"`
}

type EmulationSetScriptExecutionDisabledResult struct {
}

type EmulationSetTouchEmulationEnabledParams struct {
	SessionID string `json:"-"`
	// Whether the touch event emulation should be enabled.
	Enabled bool `json:"enabled"`
	// Maximum touch points supported. Defaults to one.
	MaxTouchPoints *int `json:"maxTouchPoints,omitempty"`
}

type EmulationSetTouchEmulationEnabledResult struct {
}

type EmulationSetVirtualTimePolicyParams struct {
	SessionID string                     `json:"-"`
	Policy    EmulationVirtualTimePolicy `json:"policy"`
	// If set, after this many virtual milliseconds have elapsed virtual time will be paused and a
	Budget *float64 `json:"budget,omitempty"`
	// If set this specifies the maximum number of tasks that can be run before virtual is forced
	MaxVirtualTimeTaskStarvationCount *int `json:"maxVirtualTimeTaskStarvationCount,omitempty"`
	// If set, base::Time::Now will be overridden to initially return this value.
	InitialVirtualTime *NetworkTimeSinceEpoch `json:"initialVirtualTime,omitempty"`
}

type EmulationSetVirtualTimePolicyResult struct {
	// Absolute timestamp at which virtual time was first enabled (up time in milliseconds).
	VirtualTimeTicksBase float64 `json:"virtualTimeTicksBase"`
}

type EmulationSetLocaleOverrideParams struct {
	SessionID string `json:"-"`
	// ICU style C locale (e.g. "en_US"). If not specified or empty, disables the override and
	Locale *string `json:"locale,omitempty"`
}

type EmulationSetLocaleOverrideResult struct {
}

type EmulationSetTimezoneOverrideParams struct {
	SessionID string `json:"-"`
	// The timezone identifier. List of supported timezones:
	TimezoneID string `json:"timezoneId"`
}

type EmulationSetTimezoneOverrideResult struct {
}

type EmulationSetVisibleSizeParams struct {
	SessionID string `json:"-"`
	// Frame width (DIP).
	Width int `json:"width"`
	// Frame height (DIP).
	Height int `json:"height"`
}

type EmulationSetVisibleSizeResult struct {
}

type EmulationSetDisabledImageTypesParams struct {
	SessionID string `json:"-"`
	// Image types to disable.
	ImageTypes []EmulationDisabledImageType `json:"imageTypes"`
}

type EmulationSetDisabledImageTypesResult struct {
}

type EmulationSetDataSaverOverrideParams struct {
	SessionID string `json:"-"`
	// Override value. Omitting the parameter disables the override.
	DataSaverEnabled *bool `json:"dataSaverEnabled,omitempty"`
}

type EmulationSetDataSaverOverrideResult struct {
}

type EmulationSetHardwareConcurrencyOverrideParams struct {
	SessionID string `json:"-"`
	// Hardware concurrency to report
	HardwareConcurrency int `json:"hardwareConcurrency"`
}

type EmulationSetHardwareConcurrencyOverrideResult struct {
}

type EmulationSetUserAgentOverrideParams struct {
	SessionID string `json:"-"`
	// User agent to use.
	UserAgent string `json:"userAgent"`
	// Browser language to emulate.
	AcceptLanguage *string `json:"acceptLanguage,omitempty"`
	// The platform navigator.platform should return.
	Platform *string `json:"platform,omitempty"`
	// To be sent in Sec-CH-UA-* headers and returned in navigator.userAgentData
	UserAgentMetadata *EmulationUserAgentMetadata `json:"userAgentMetadata,omitempty"`
}

type EmulationSetUserAgentOverrideResult struct {
}

type EmulationSetAutomationOverrideParams struct {
	SessionID string `json:"-"`
	// Whether the override should be enabled.
	Enabled bool `json:"enabled"`
}

type EmulationSetAutomationOverrideResult struct {
}

type EmulationSetSmallViewportHeightDifferenceOverrideParams struct {
	SessionID string `json:"-"`
	// This will cause an element of size 100svh to be `difference` pixels smaller than an element
	Difference int `json:"difference"`
}

type EmulationSetSmallViewportHeightDifferenceOverrideResult struct {
}

type EmulationGetScreenInfosParams struct {
	SessionID string `json:"-"`
}

type EmulationGetScreenInfosResult struct {
	ScreenInfos []EmulationScreenInfo `json:"screenInfos"`
}

type EmulationAddScreenParams struct {
	SessionID string `json:"-"`
	// Offset of the left edge of the screen in pixels.
	Left int `json:"left"`
	// Offset of the top edge of the screen in pixels.
	Top int `json:"top"`
	// The width of the screen in pixels.
	Width int `json:"width"`
	// The height of the screen in pixels.
	Height int `json:"height"`
	// Specifies the screen's work area. Default is entire screen.
	WorkAreaInsets *EmulationWorkAreaInsets `json:"workAreaInsets,omitempty"`
	// Specifies the screen's device pixel ratio. Default is 1.
	DevicePixelRatio *float64 `json:"devicePixelRatio,omitempty"`
	// Specifies the screen's rotation angle. Available values are 0, 90, 180 and 270. Default is 0.
	Rotation *int `json:"rotation,omitempty"`
	// Specifies the screen's color depth in bits. Default is 24.
	ColorDepth *int `json:"colorDepth,omitempty"`
	// Specifies the descriptive label for the screen. Default is none.
	Label *string `json:"label,omitempty"`
	// Indicates whether the screen is internal to the device or external, attached to the device. Default is false.
	IsInternal *bool `json:"isInternal,omitempty"`
}

type EmulationAddScreenResult struct {
	ScreenInfo EmulationScreenInfo `json:"screenInfo"`
}

type EmulationUpdateScreenParams struct {
	SessionID string `json:"-"`
	// Target screen identifier.
	ScreenID EmulationScreenID `json:"screenId"`
	// Offset of the left edge of the screen in pixels.
	Left *int `json:"left,omitempty"`
	// Offset of the top edge of the screen in pixels.
	Top *int `json:"top,omitempty"`
	// The width of the screen in pixels.
	Width *int `json:"width,omitempty"`
	// The height of the screen in pixels.
	Height *int `json:"height,omitempty"`
	// Specifies the screen's work area.
	WorkAreaInsets *EmulationWorkAreaInsets `json:"workAreaInsets,omitempty"`
	// Specifies the screen's device pixel ratio.
	DevicePixelRatio *float64 `json:"devicePixelRatio,omitempty"`
	// Specifies the screen's rotation angle. Available values are 0, 90, 180 and 270.
	Rotation *int `json:"rotation,omitempty"`
	// Specifies the screen's color depth in bits.
	ColorDepth *int `json:"colorDepth,omitempty"`
	// Specifies the descriptive label for the screen.
	Label *string `json:"label,omitempty"`
	// Indicates whether the screen is internal to the device or external, attached to the device. Default is false.
	IsInternal *bool `json:"isInternal,omitempty"`
}

type EmulationUpdateScreenResult struct {
	ScreenInfo EmulationScreenInfo `json:"screenInfo"`
}

type EmulationRemoveScreenParams struct {
	SessionID string            `json:"-"`
	ScreenID  EmulationScreenID `json:"screenId"`
}

type EmulationRemoveScreenResult struct {
}

type EmulationSetPrimaryScreenParams struct {
	SessionID string            `json:"-"`
	ScreenID  EmulationScreenID `json:"screenId"`
}

type EmulationSetPrimaryScreenResult struct {
}

type EmulationVirtualTimeBudgetExpiredEvent struct {
}

type EmulationScreenOrientationLockChangedEvent struct {
	// Whether the screen orientation is currently locked.
	Locked bool `json:"locked"`
	// The orientation lock type requested by the page. Only set when locked is true.
	Orientation *EmulationScreenOrientation `json:"orientation,omitempty"`
}

type EventBreakpointsSetInstrumentationBreakpointParams struct {
	SessionID string `json:"-"`
	// Instrumentation name to stop on.
	EventName string `json:"eventName"`
}

type EventBreakpointsSetInstrumentationBreakpointResult struct {
}

type EventBreakpointsRemoveInstrumentationBreakpointParams struct {
	SessionID string `json:"-"`
	// Instrumentation name to stop on.
	EventName string `json:"eventName"`
}

type EventBreakpointsRemoveInstrumentationBreakpointResult struct {
}

type EventBreakpointsDisableParams struct {
	SessionID string `json:"-"`
}

type EventBreakpointsDisableResult struct {
}

type ExtensionsStorageArea string

type ExtensionsExtensionInfo struct {
	// Extension id.
	ID string `json:"id"`
	// Extension name.
	Name string `json:"name"`
	// Extension version.
	Version string `json:"version"`
	// The path from which the extension was loaded.
	Path string `json:"path"`
	// Extension enabled status.
	Enabled bool `json:"enabled"`
}

type ExtensionsTriggerActionParams struct {
	SessionID string `json:"-"`
	// Extension id.
	ID string `json:"id"`
	// A tab target ID to trigger the default extension action on.
	TargetID string `json:"targetId"`
}

type ExtensionsTriggerActionResult struct {
}

type ExtensionsLoadUnpackedParams struct {
	SessionID string `json:"-"`
	// Absolute file path.
	Path string `json:"path"`
	// Enable the extension in incognito
	EnableInIncognito *bool `json:"enableInIncognito,omitempty"`
}

type ExtensionsLoadUnpackedResult struct {
	// Extension id.
	ID string `json:"id"`
}

type ExtensionsGetExtensionsParams struct {
	SessionID string `json:"-"`
}

type ExtensionsGetExtensionsResult struct {
	Extensions []ExtensionsExtensionInfo `json:"extensions"`
}

type ExtensionsUninstallParams struct {
	SessionID string `json:"-"`
	// Extension id.
	ID string `json:"id"`
}

type ExtensionsUninstallResult struct {
}

type ExtensionsGetStorageItemsParams struct {
	SessionID string `json:"-"`
	// ID of extension.
	ID string `json:"id"`
	// StorageArea to retrieve data from.
	StorageArea ExtensionsStorageArea `json:"storageArea"`
	// Keys to retrieve.
	Keys []string `json:"keys,omitempty"`
}

type ExtensionsGetStorageItemsResult struct {
	Data map[string]any `json:"data"`
}

type ExtensionsRemoveStorageItemsParams struct {
	SessionID string `json:"-"`
	// ID of extension.
	ID string `json:"id"`
	// StorageArea to remove data from.
	StorageArea ExtensionsStorageArea `json:"storageArea"`
	// Keys to remove.
	Keys []string `json:"keys"`
}

type ExtensionsRemoveStorageItemsResult struct {
}

type ExtensionsClearStorageItemsParams struct {
	SessionID string `json:"-"`
	// ID of extension.
	ID string `json:"id"`
	// StorageArea to remove data from.
	StorageArea ExtensionsStorageArea `json:"storageArea"`
}

type ExtensionsClearStorageItemsResult struct {
}

type ExtensionsSetStorageItemsParams struct {
	SessionID string `json:"-"`
	// ID of extension.
	ID string `json:"id"`
	// StorageArea to set data in.
	StorageArea ExtensionsStorageArea `json:"storageArea"`
	// Values to set.
	Values map[string]any `json:"values"`
}

type ExtensionsSetStorageItemsResult struct {
}

type FedCmLoginState string

type FedCmDialogType string

type FedCmDialogButton string

type FedCmAccountURLType string

type FedCmAccount struct {
	AccountID    string          `json:"accountId"`
	Email        string          `json:"email"`
	Name         string          `json:"name"`
	GivenName    string          `json:"givenName"`
	PictureURL   string          `json:"pictureUrl"`
	IdpConfigURL string          `json:"idpConfigUrl"`
	IdpLoginURL  string          `json:"idpLoginUrl"`
	LoginState   FedCmLoginState `json:"loginState"`
	// These two are only set if the loginState is signUp
	TermsOfServiceURL *string `json:"termsOfServiceUrl,omitempty"`
	PrivacyPolicyURL  *string `json:"privacyPolicyUrl,omitempty"`
}

type FedCmEnableParams struct {
	SessionID string `json:"-"`
	// Allows callers to disable the promise rejection delay that would
	DisableRejectionDelay *bool `json:"disableRejectionDelay,omitempty"`
}

type FedCmEnableResult struct {
}

type FedCmDisableParams struct {
	SessionID string `json:"-"`
}

type FedCmDisableResult struct {
}

type FedCmSelectAccountParams struct {
	SessionID    string `json:"-"`
	DialogID     string `json:"dialogId"`
	AccountIndex int    `json:"accountIndex"`
}

type FedCmSelectAccountResult struct {
}

type FedCmClickDialogButtonParams struct {
	SessionID    string            `json:"-"`
	DialogID     string            `json:"dialogId"`
	DialogButton FedCmDialogButton `json:"dialogButton"`
}

type FedCmClickDialogButtonResult struct {
}

type FedCmOpenURLParams struct {
	SessionID      string              `json:"-"`
	DialogID       string              `json:"dialogId"`
	AccountIndex   int                 `json:"accountIndex"`
	AccountURLType FedCmAccountURLType `json:"accountUrlType"`
}

type FedCmOpenURLResult struct {
}

type FedCmDismissDialogParams struct {
	SessionID       string `json:"-"`
	DialogID        string `json:"dialogId"`
	TriggerCooldown *bool  `json:"triggerCooldown,omitempty"`
}

type FedCmDismissDialogResult struct {
}

type FedCmResetCooldownParams struct {
	SessionID string `json:"-"`
}

type FedCmResetCooldownResult struct {
}

type FedCmDialogShownEvent struct {
	DialogID   string          `json:"dialogId"`
	DialogType FedCmDialogType `json:"dialogType"`
	Accounts   []FedCmAccount  `json:"accounts"`
	// These exist primarily so that the caller can verify the
	Title    string  `json:"title"`
	Subtitle *string `json:"subtitle,omitempty"`
}

type FedCmDialogClosedEvent struct {
	DialogID string `json:"dialogId"`
}

type FetchRequestID string

type FetchRequestStage string

type FetchRequestPattern struct {
	// Wildcards (`'*'` -> zero or more, `'?'` -> exactly one) are allowed. Escape character is
	URLPattern *string `json:"urlPattern,omitempty"`
	// If set, only requests for matching resource types will be intercepted.
	ResourceType *NetworkResourceType `json:"resourceType,omitempty"`
	// Stage at which to begin intercepting requests. Default is Request.
	RequestStage *FetchRequestStage `json:"requestStage,omitempty"`
}

type FetchHeaderEntry struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type FetchAuthChallenge struct {
	// Source of the authentication challenge.
	Source *string `json:"source,omitempty"`
	// Origin of the challenger.
	Origin string `json:"origin"`
	// The authentication scheme used, such as basic or digest
	Scheme string `json:"scheme"`
	// The realm of the challenge. May be empty.
	Realm string `json:"realm"`
}

type FetchAuthChallengeResponse struct {
	// The decision on what to do in response to the authorization challenge.  Default means
	Response string `json:"response"`
	// The username to provide, possibly empty. Should only be set if response is
	Username *string `json:"username,omitempty"`
	// The password to provide, possibly empty. Should only be set if response is
	Password *string `json:"password,omitempty"`
}

type FetchDisableParams struct {
	SessionID string `json:"-"`
}

type FetchDisableResult struct {
}

type FetchEnableParams struct {
	SessionID string `json:"-"`
	// If specified, only requests matching any of these patterns will produce
	Patterns []FetchRequestPattern `json:"patterns,omitempty"`
	// If true, authRequired events will be issued and requests will be paused
	HandleAuthRequests *bool `json:"handleAuthRequests,omitempty"`
}

type FetchEnableResult struct {
}

type FetchFailRequestParams struct {
	SessionID string `json:"-"`
	// An id the client received in requestPaused event.
	RequestID FetchRequestID `json:"requestId"`
	// Causes the request to fail with the given reason.
	ErrorReason NetworkErrorReason `json:"errorReason"`
}

type FetchFailRequestResult struct {
}

type FetchFulfillRequestParams struct {
	SessionID string `json:"-"`
	// An id the client received in requestPaused event.
	RequestID FetchRequestID `json:"requestId"`
	// An HTTP response code.
	ResponseCode int `json:"responseCode"`
	// Response headers.
	ResponseHeaders []FetchHeaderEntry `json:"responseHeaders,omitempty"`
	// Alternative way of specifying response headers as a \0-separated
	BinaryResponseHeaders *string `json:"binaryResponseHeaders,omitempty"`
	// A response body. If absent, original response body will be used if
	Body *string `json:"body,omitempty"`
	// A textual representation of responseCode.
	ResponsePhrase *string `json:"responsePhrase,omitempty"`
}

type FetchFulfillRequestResult struct {
}

type FetchContinueRequestParams struct {
	SessionID string `json:"-"`
	// An id the client received in requestPaused event.
	RequestID FetchRequestID `json:"requestId"`
	// If set, the request url will be modified in a way that's not observable by page.
	URL *string `json:"url,omitempty"`
	// If set, the request method is overridden.
	Method *string `json:"method,omitempty"`
	// If set, overrides the post data in the request. (Encoded as a base64 string when passed over JSON)
	PostData *string `json:"postData,omitempty"`
	// If set, overrides the request headers. Note that the overrides do not
	Headers []FetchHeaderEntry `json:"headers,omitempty"`
	// If set, overrides response interception behavior for this request.
	InterceptResponse *bool `json:"interceptResponse,omitempty"`
}

type FetchContinueRequestResult struct {
}

type FetchContinueWithAuthParams struct {
	SessionID string `json:"-"`
	// An id the client received in authRequired event.
	RequestID FetchRequestID `json:"requestId"`
	// Response to  with an authChallenge.
	AuthChallengeResponse FetchAuthChallengeResponse `json:"authChallengeResponse"`
}

type FetchContinueWithAuthResult struct {
}

type FetchContinueResponseParams struct {
	SessionID string `json:"-"`
	// An id the client received in requestPaused event.
	RequestID FetchRequestID `json:"requestId"`
	// An HTTP response code. If absent, original response code will be used.
	ResponseCode *int `json:"responseCode,omitempty"`
	// A textual representation of responseCode.
	ResponsePhrase *string `json:"responsePhrase,omitempty"`
	// Response headers. If absent, original response headers will be used.
	ResponseHeaders []FetchHeaderEntry `json:"responseHeaders,omitempty"`
	// Alternative way of specifying response headers as a \0-separated
	BinaryResponseHeaders *string `json:"binaryResponseHeaders,omitempty"`
}

type FetchContinueResponseResult struct {
}

type FetchGetResponseBodyParams struct {
	SessionID string `json:"-"`
	// Identifier for the intercepted request to get body for.
	RequestID FetchRequestID `json:"requestId"`
}

type FetchGetResponseBodyResult struct {
	// Response body.
	Body string `json:"body"`
	// True, if content was sent as base64.
	Base64Encoded bool `json:"base64Encoded"`
}

type FetchTakeResponseBodyAsStreamParams struct {
	SessionID string         `json:"-"`
	RequestID FetchRequestID `json:"requestId"`
}

type FetchTakeResponseBodyAsStreamResult struct {
	Stream IOStreamHandle `json:"stream"`
}

type FetchRequestPausedEvent struct {
	// Each request the page makes will have a unique id.
	RequestID FetchRequestID `json:"requestId"`
	// The details of the request.
	Request NetworkRequest `json:"request"`
	// The id of the frame that initiated the request.
	FrameID PageFrameID `json:"frameId"`
	// How the requested resource will be used.
	ResourceType NetworkResourceType `json:"resourceType"`
	// Response error if intercepted at response stage.
	ResponseErrorReason *NetworkErrorReason `json:"responseErrorReason,omitempty"`
	// Response code if intercepted at response stage.
	ResponseStatusCode *int `json:"responseStatusCode,omitempty"`
	// Response status text if intercepted at response stage.
	ResponseStatusText *string `json:"responseStatusText,omitempty"`
	// Response headers if intercepted at the response stage.
	ResponseHeaders []FetchHeaderEntry `json:"responseHeaders,omitempty"`
	// If the intercepted request had a corresponding Network.requestWillBeSent event fired for it,
	NetworkID *NetworkRequestID `json:"networkId,omitempty"`
	// If the request is due to a redirect response from the server, the id of the request that
	RedirectedRequestID *FetchRequestID `json:"redirectedRequestId,omitempty"`
}

type FetchAuthRequiredEvent struct {
	// Each request the page makes will have a unique id.
	RequestID FetchRequestID `json:"requestId"`
	// The details of the request.
	Request NetworkRequest `json:"request"`
	// The id of the frame that initiated the request.
	FrameID PageFrameID `json:"frameId"`
	// How the requested resource will be used.
	ResourceType NetworkResourceType `json:"resourceType"`
	// Details of the Authorization Challenge encountered.
	AuthChallenge FetchAuthChallenge `json:"authChallenge"`
}

type FileSystemFile struct {
	Name string `json:"name"`
	// Timestamp
	LastModified NetworkTimeSinceEpoch `json:"lastModified"`
	// Size in bytes
	Size float64 `json:"size"`
	Type string  `json:"type"`
}

type FileSystemDirectory struct {
	Name              string   `json:"name"`
	NestedDirectories []string `json:"nestedDirectories"`
	// Files that are directly nested under this directory.
	NestedFiles []FileSystemFile `json:"nestedFiles"`
}

type FileSystemBucketFileSystemLocator struct {
	// Storage key
	StorageKey StorageSerializedStorageKey `json:"storageKey"`
	// Bucket name. Not passing a `bucketName` will retrieve the default Bucket. (https://developer.mozilla.org/en-US/docs/Web/API/Storage_API#storage_buckets)
	BucketName *string `json:"bucketName,omitempty"`
	// Path to the directory using each path component as an array item.
	PathComponents []string `json:"pathComponents"`
}

type FileSystemGetDirectoryParams struct {
	SessionID               string                            `json:"-"`
	BucketFileSystemLocator FileSystemBucketFileSystemLocator `json:"bucketFileSystemLocator"`
}

type FileSystemGetDirectoryResult struct {
	// Returns the directory object at the path.
	Directory FileSystemDirectory `json:"directory"`
}

type HeadlessExperimentalScreenshotParams struct {
	// Image compression format (defaults to png).
	Format *string `json:"format,omitempty"`
	// Compression quality from range [0..100] (jpeg and webp only).
	Quality *int `json:"quality,omitempty"`
	// Optimize image encoding for speed, not for resulting size (defaults to false)
	OptimizeForSpeed *bool `json:"optimizeForSpeed,omitempty"`
}

type HeadlessExperimentalBeginFrameParams struct {
	SessionID string `json:"-"`
	// Timestamp of this BeginFrame in Renderer TimeTicks (milliseconds of uptime). If not set,
	FrameTimeTicks *float64 `json:"frameTimeTicks,omitempty"`
	// The interval between BeginFrames that is reported to the compositor, in milliseconds.
	Interval *float64 `json:"interval,omitempty"`
	// Whether updates should not be committed and drawn onto the display. False by default. If
	NoDisplayUpdates *bool `json:"noDisplayUpdates,omitempty"`
	// If set, a screenshot of the frame will be captured and returned in the response. Otherwise,
	Screenshot *HeadlessExperimentalScreenshotParams `json:"screenshot,omitempty"`
}

type HeadlessExperimentalBeginFrameResult struct {
	// Whether the BeginFrame resulted in damage and, thus, a new frame was committed to the
	HasDamage bool `json:"hasDamage"`
	// Base64-encoded image data of the screenshot, if one was requested and successfully taken. (Encoded as a base64 string when passed over JSON)
	ScreenshotData *string `json:"screenshotData,omitempty"`
}

type HeadlessExperimentalDisableParams struct {
	SessionID string `json:"-"`
}

type HeadlessExperimentalDisableResult struct {
}

type HeadlessExperimentalEnableParams struct {
	SessionID string `json:"-"`
}

type HeadlessExperimentalEnableResult struct {
}

type HeapProfilerHeapSnapshotObjectID string

type HeapProfilerSamplingHeapProfileNode struct {
	// Function location.
	CallFrame RuntimeCallFrame `json:"callFrame"`
	// Allocations size in bytes for the node excluding children.
	SelfSize float64 `json:"selfSize"`
	// Node id. Ids are unique across all profiles collected between startSampling and stopSampling.
	ID int `json:"id"`
	// Child nodes.
	Children []HeapProfilerSamplingHeapProfileNode `json:"children"`
}

type HeapProfilerSamplingHeapProfileSample struct {
	// Allocation size in bytes attributed to the sample.
	Size float64 `json:"size"`
	// Id of the corresponding profile tree node.
	NodeID int `json:"nodeId"`
	// Time-ordered sample ordinal number. It is unique across all profiles retrieved
	Ordinal float64 `json:"ordinal"`
}

type HeapProfilerSamplingHeapProfile struct {
	Head    HeapProfilerSamplingHeapProfileNode     `json:"head"`
	Samples []HeapProfilerSamplingHeapProfileSample `json:"samples"`
}

type HeapProfilerAddInspectedHeapObjectParams struct {
	SessionID string `json:"-"`
	// Heap snapshot object id to be accessible by means of $x command line API.
	HeapObjectID HeapProfilerHeapSnapshotObjectID `json:"heapObjectId"`
}

type HeapProfilerAddInspectedHeapObjectResult struct {
}

type HeapProfilerCollectGarbageParams struct {
	SessionID string `json:"-"`
}

type HeapProfilerCollectGarbageResult struct {
}

type HeapProfilerDisableParams struct {
	SessionID string `json:"-"`
}

type HeapProfilerDisableResult struct {
}

type HeapProfilerEnableParams struct {
	SessionID string `json:"-"`
}

type HeapProfilerEnableResult struct {
}

type HeapProfilerGetHeapObjectIDParams struct {
	SessionID string `json:"-"`
	// Identifier of the object to get heap object id for.
	ObjectID RuntimeRemoteObjectID `json:"objectId"`
}

type HeapProfilerGetHeapObjectIDResult struct {
	// Id of the heap snapshot object corresponding to the passed remote object id.
	HeapSnapshotObjectID HeapProfilerHeapSnapshotObjectID `json:"heapSnapshotObjectId"`
}

type HeapProfilerGetObjectByHeapObjectIDParams struct {
	SessionID string                           `json:"-"`
	ObjectID  HeapProfilerHeapSnapshotObjectID `json:"objectId"`
	// Symbolic group name that can be used to release multiple objects.
	ObjectGroup *string `json:"objectGroup,omitempty"`
}

type HeapProfilerGetObjectByHeapObjectIDResult struct {
	// Evaluation result.
	Result RuntimeRemoteObject `json:"result"`
}

type HeapProfilerGetSamplingProfileParams struct {
	SessionID string `json:"-"`
}

type HeapProfilerGetSamplingProfileResult struct {
	// Return the sampling profile being collected.
	Profile HeapProfilerSamplingHeapProfile `json:"profile"`
}

type HeapProfilerStartSamplingParams struct {
	SessionID string `json:"-"`
	// Average sample interval in bytes. Poisson distribution is used for the intervals. The
	SamplingInterval *float64 `json:"samplingInterval,omitempty"`
	// Maximum stack depth. The default value is 128.
	StackDepth *float64 `json:"stackDepth,omitempty"`
	// By default, the sampling heap profiler reports only objects which are
	IncludeObjectsCollectedByMajorGC *bool `json:"includeObjectsCollectedByMajorGC,omitempty"`
	// By default, the sampling heap profiler reports only objects which are
	IncludeObjectsCollectedByMinorGC *bool `json:"includeObjectsCollectedByMinorGC,omitempty"`
}

type HeapProfilerStartSamplingResult struct {
}

type HeapProfilerStartTrackingHeapObjectsParams struct {
	SessionID        string `json:"-"`
	TrackAllocations *bool  `json:"trackAllocations,omitempty"`
}

type HeapProfilerStartTrackingHeapObjectsResult struct {
}

type HeapProfilerStopSamplingParams struct {
	SessionID string `json:"-"`
}

type HeapProfilerStopSamplingResult struct {
	// Recorded sampling heap profile.
	Profile HeapProfilerSamplingHeapProfile `json:"profile"`
}

type HeapProfilerStopTrackingHeapObjectsParams struct {
	SessionID string `json:"-"`
	// If true 'reportHeapSnapshotProgress' events will be generated while snapshot is being taken
	ReportProgress *bool `json:"reportProgress,omitempty"`
	// Deprecated in favor of `exposeInternals`.
	TreatGlobalObjectsAsRoots *bool `json:"treatGlobalObjectsAsRoots,omitempty"`
	// If true, numerical values are included in the snapshot
	CaptureNumericValue *bool `json:"captureNumericValue,omitempty"`
	// If true, exposes internals of the snapshot.
	ExposeInternals *bool `json:"exposeInternals,omitempty"`
}

type HeapProfilerStopTrackingHeapObjectsResult struct {
}

type HeapProfilerTakeHeapSnapshotParams struct {
	SessionID string `json:"-"`
	// If true 'reportHeapSnapshotProgress' events will be generated while snapshot is being taken.
	ReportProgress *bool `json:"reportProgress,omitempty"`
	// If true, a raw snapshot without artificial roots will be generated.
	TreatGlobalObjectsAsRoots *bool `json:"treatGlobalObjectsAsRoots,omitempty"`
	// If true, numerical values are included in the snapshot
	CaptureNumericValue *bool `json:"captureNumericValue,omitempty"`
	// If true, exposes internals of the snapshot.
	ExposeInternals *bool `json:"exposeInternals,omitempty"`
}

type HeapProfilerTakeHeapSnapshotResult struct {
}

type HeapProfilerAddHeapSnapshotChunkEvent struct {
	Chunk string `json:"chunk"`
}

type HeapProfilerHeapStatsUpdateEvent struct {
	// An array of triplets. Each triplet describes a fragment. The first integer is the fragment
	StatsUpdate []int `json:"statsUpdate"`
}

type HeapProfilerLastSeenObjectIDEvent struct {
	LastSeenObjectID int     `json:"lastSeenObjectId"`
	Timestamp        float64 `json:"timestamp"`
}

type HeapProfilerReportHeapSnapshotProgressEvent struct {
	Done     int   `json:"done"`
	Total    int   `json:"total"`
	Finished *bool `json:"finished,omitempty"`
}

type HeapProfilerResetProfilesEvent struct {
}

type IOStreamHandle string

type IOCloseParams struct {
	SessionID string `json:"-"`
	// Handle of the stream to close.
	Handle IOStreamHandle `json:"handle"`
}

type IOCloseResult struct {
}

type IOReadParams struct {
	SessionID string `json:"-"`
	// Handle of the stream to read.
	Handle IOStreamHandle `json:"handle"`
	// Seek to the specified offset before reading (if not specified, proceed with offset
	Offset *int `json:"offset,omitempty"`
	// Maximum number of bytes to read (left upon the agent discretion if not specified).
	Size *int `json:"size,omitempty"`
}

type IOReadResult struct {
	// Set if the data is base64-encoded
	Base64Encoded *bool `json:"base64Encoded,omitempty"`
	// Data that were read.
	Data string `json:"data"`
	// Set if the end-of-file condition occurred while reading.
	Eof bool `json:"eof"`
}

type IOResolveBlobParams struct {
	SessionID string `json:"-"`
	// Object id of a Blob object wrapper.
	ObjectID RuntimeRemoteObjectID `json:"objectId"`
}

type IOResolveBlobResult struct {
	// UUID of the specified Blob.
	UUID string `json:"uuid"`
}

type IndexedDBDatabaseWithObjectStores struct {
	// Database name.
	Name string `json:"name"`
	// Database version (type is not 'integer', as the standard
	Version float64 `json:"version"`
	// Object stores in this database.
	ObjectStores []IndexedDBObjectStore `json:"objectStores"`
}

type IndexedDBObjectStore struct {
	// Object store name.
	Name string `json:"name"`
	// Object store key path.
	KeyPath IndexedDBKeyPath `json:"keyPath"`
	// If true, object store has auto increment flag set.
	AutoIncrement bool `json:"autoIncrement"`
	// Indexes in this object store.
	Indexes []IndexedDBObjectStoreIndex `json:"indexes"`
}

type IndexedDBObjectStoreIndex struct {
	// Index name.
	Name string `json:"name"`
	// Index key path.
	KeyPath IndexedDBKeyPath `json:"keyPath"`
	// If true, index is unique.
	Unique bool `json:"unique"`
	// If true, index allows multiple entries for a key.
	MultiEntry bool `json:"multiEntry"`
}

type IndexedDBKey struct {
	// Key type.
	Type string `json:"type"`
	// Number value.
	Number *float64 `json:"number,omitempty"`
	// String value.
	String *string `json:"string,omitempty"`
	// Date value.
	Date *float64 `json:"date,omitempty"`
	// Array value.
	Array []IndexedDBKey `json:"array,omitempty"`
}

type IndexedDBKeyRange struct {
	// Lower bound.
	Lower *IndexedDBKey `json:"lower,omitempty"`
	// Upper bound.
	Upper *IndexedDBKey `json:"upper,omitempty"`
	// If true lower bound is open.
	LowerOpen bool `json:"lowerOpen"`
	// If true upper bound is open.
	UpperOpen bool `json:"upperOpen"`
}

type IndexedDBDataEntry struct {
	// Key object.
	Key RuntimeRemoteObject `json:"key"`
	// Primary key object.
	PrimaryKey RuntimeRemoteObject `json:"primaryKey"`
	// Value object.
	Value RuntimeRemoteObject `json:"value"`
}

type IndexedDBKeyPath struct {
	// Key path type.
	Type string `json:"type"`
	// String value.
	String *string `json:"string,omitempty"`
	// Array value.
	Array []string `json:"array,omitempty"`
}

type IndexedDBClearObjectStoreParams struct {
	SessionID string `json:"-"`
	// At least and at most one of securityOrigin, storageKey, or storageBucket must be specified.
	SecurityOrigin *string `json:"securityOrigin,omitempty"`
	// Storage key.
	StorageKey *string `json:"storageKey,omitempty"`
	// Storage bucket. If not specified, it uses the default bucket.
	StorageBucket *StorageStorageBucket `json:"storageBucket,omitempty"`
	// Database name.
	DatabaseName string `json:"databaseName"`
	// Object store name.
	ObjectStoreName string `json:"objectStoreName"`
}

type IndexedDBClearObjectStoreResult struct {
}

type IndexedDBDeleteDatabaseParams struct {
	SessionID string `json:"-"`
	// At least and at most one of securityOrigin, storageKey, or storageBucket must be specified.
	SecurityOrigin *string `json:"securityOrigin,omitempty"`
	// Storage key.
	StorageKey *string `json:"storageKey,omitempty"`
	// Storage bucket. If not specified, it uses the default bucket.
	StorageBucket *StorageStorageBucket `json:"storageBucket,omitempty"`
	// Database name.
	DatabaseName string `json:"databaseName"`
}

type IndexedDBDeleteDatabaseResult struct {
}

type IndexedDBDeleteObjectStoreEntriesParams struct {
	SessionID string `json:"-"`
	// At least and at most one of securityOrigin, storageKey, or storageBucket must be specified.
	SecurityOrigin *string `json:"securityOrigin,omitempty"`
	// Storage key.
	StorageKey *string `json:"storageKey,omitempty"`
	// Storage bucket. If not specified, it uses the default bucket.
	StorageBucket   *StorageStorageBucket `json:"storageBucket,omitempty"`
	DatabaseName    string                `json:"databaseName"`
	ObjectStoreName string                `json:"objectStoreName"`
	// Range of entry keys to delete
	KeyRange IndexedDBKeyRange `json:"keyRange"`
}

type IndexedDBDeleteObjectStoreEntriesResult struct {
}

type IndexedDBDisableParams struct {
	SessionID string `json:"-"`
}

type IndexedDBDisableResult struct {
}

type IndexedDBEnableParams struct {
	SessionID string `json:"-"`
}

type IndexedDBEnableResult struct {
}

type IndexedDBRequestDataParams struct {
	SessionID string `json:"-"`
	// At least and at most one of securityOrigin, storageKey, or storageBucket must be specified.
	SecurityOrigin *string `json:"securityOrigin,omitempty"`
	// Storage key.
	StorageKey *string `json:"storageKey,omitempty"`
	// Storage bucket. If not specified, it uses the default bucket.
	StorageBucket *StorageStorageBucket `json:"storageBucket,omitempty"`
	// Database name.
	DatabaseName string `json:"databaseName"`
	// Object store name.
	ObjectStoreName string `json:"objectStoreName"`
	// Index name. If not specified, it performs an object store data request.
	IndexName *string `json:"indexName,omitempty"`
	// Number of records to skip.
	SkipCount int `json:"skipCount"`
	// Number of records to fetch.
	PageSize int `json:"pageSize"`
	// Key range.
	KeyRange *IndexedDBKeyRange `json:"keyRange,omitempty"`
}

type IndexedDBRequestDataResult struct {
	// Array of object store data entries.
	ObjectStoreDataEntries []IndexedDBDataEntry `json:"objectStoreDataEntries"`
	// If true, there are more entries to fetch in the given range.
	HasMore bool `json:"hasMore"`
}

type IndexedDBGetMetadataParams struct {
	SessionID string `json:"-"`
	// At least and at most one of securityOrigin, storageKey, or storageBucket must be specified.
	SecurityOrigin *string `json:"securityOrigin,omitempty"`
	// Storage key.
	StorageKey *string `json:"storageKey,omitempty"`
	// Storage bucket. If not specified, it uses the default bucket.
	StorageBucket *StorageStorageBucket `json:"storageBucket,omitempty"`
	// Database name.
	DatabaseName string `json:"databaseName"`
	// Object store name.
	ObjectStoreName string `json:"objectStoreName"`
}

type IndexedDBGetMetadataResult struct {
	// the entries count
	EntriesCount float64 `json:"entriesCount"`
	// the current value of key generator, to become the next inserted
	KeyGeneratorValue float64 `json:"keyGeneratorValue"`
}

type IndexedDBRequestDatabaseParams struct {
	SessionID string `json:"-"`
	// At least and at most one of securityOrigin, storageKey, or storageBucket must be specified.
	SecurityOrigin *string `json:"securityOrigin,omitempty"`
	// Storage key.
	StorageKey *string `json:"storageKey,omitempty"`
	// Storage bucket. If not specified, it uses the default bucket.
	StorageBucket *StorageStorageBucket `json:"storageBucket,omitempty"`
	// Database name.
	DatabaseName string `json:"databaseName"`
}

type IndexedDBRequestDatabaseResult struct {
	// Database with an array of object stores.
	DatabaseWithObjectStores IndexedDBDatabaseWithObjectStores `json:"databaseWithObjectStores"`
}

type IndexedDBRequestDatabaseNamesParams struct {
	SessionID string `json:"-"`
	// At least and at most one of securityOrigin, storageKey, or storageBucket must be specified.
	SecurityOrigin *string `json:"securityOrigin,omitempty"`
	// Storage key.
	StorageKey *string `json:"storageKey,omitempty"`
	// Storage bucket. If not specified, it uses the default bucket.
	StorageBucket *StorageStorageBucket `json:"storageBucket,omitempty"`
}

type IndexedDBRequestDatabaseNamesResult struct {
	// Database names for origin.
	DatabaseNames []string `json:"databaseNames"`
}

type InputTouchPoint struct {
	// X coordinate of the event relative to the main frame's viewport in CSS pixels.
	X float64 `json:"x"`
	// Y coordinate of the event relative to the main frame's viewport in CSS pixels. 0 refers to
	Y float64 `json:"y"`
	// X radius of the touch area (default: 1.0).
	RadiusX *float64 `json:"radiusX,omitempty"`
	// Y radius of the touch area (default: 1.0).
	RadiusY *float64 `json:"radiusY,omitempty"`
	// Rotation angle (default: 0.0).
	RotationAngle *float64 `json:"rotationAngle,omitempty"`
	// Force (default: 1.0).
	Force *float64 `json:"force,omitempty"`
	// The normalized tangential pressure, which has a range of [-1,1] (default: 0).
	TangentialPressure *float64 `json:"tangentialPressure,omitempty"`
	// The plane angle between the Y-Z plane and the plane containing both the stylus axis and the Y axis, in degrees of the range [-90,90], a positive tiltX is to the right (default: 0)
	TiltX *float64 `json:"tiltX,omitempty"`
	// The plane angle between the X-Z plane and the plane containing both the stylus axis and the X axis, in degrees of the range [-90,90], a positive tiltY is towards the user (default:
	TiltY *float64 `json:"tiltY,omitempty"`
	// The clockwise rotation of a pen stylus around its own major axis, in degrees in the range [0,359] (default: 0).
	Twist *int `json:"twist,omitempty"`
	// Identifier used to track touch sources between events, must be unique within an event.
	ID *float64 `json:"id,omitempty"`
}

type InputGestureSourceType string

type InputMouseButton string

type InputTimeSinceEpoch float64

type InputDragDataItem struct {
	// Mime type of the dragged data.
	MimeType string `json:"mimeType"`
	// Depending of the value of `mimeType`, it contains the dragged link,
	Data string `json:"data"`
	// Title associated with a link. Only valid when `mimeType` == "text/uri-list".
	Title *string `json:"title,omitempty"`
	// Stores the base URL for the contained markup. Only valid when `mimeType`
	BaseURL *string `json:"baseURL,omitempty"`
}

type InputDragData struct {
	Items []InputDragDataItem `json:"items"`
	// List of filenames that should be included when dropping
	Files []string `json:"files,omitempty"`
	// Bit field representing allowed drag operations. Copy = 1, Link = 2, Move = 16
	DragOperationsMask int `json:"dragOperationsMask"`
}

type InputDispatchDragEventParams struct {
	SessionID string `json:"-"`
	// Type of the drag event.
	Type string `json:"type"`
	// X coordinate of the event relative to the main frame's viewport in CSS pixels.
	X float64 `json:"x"`
	// Y coordinate of the event relative to the main frame's viewport in CSS pixels. 0 refers to
	Y    float64       `json:"y"`
	Data InputDragData `json:"data"`
	// Bit field representing pressed modifier keys. Alt=1, Ctrl=2, Meta/Command=4, Shift=8
	Modifiers *int `json:"modifiers,omitempty"`
}

type InputDispatchDragEventResult struct {
}

type InputDispatchKeyEventParams struct {
	SessionID string `json:"-"`
	// Type of the key event.
	Type string `json:"type"`
	// Bit field representing pressed modifier keys. Alt=1, Ctrl=2, Meta/Command=4, Shift=8
	Modifiers *int `json:"modifiers,omitempty"`
	// Time at which the event occurred.
	Timestamp *InputTimeSinceEpoch `json:"timestamp,omitempty"`
	// Text as generated by processing a virtual key code with a keyboard layout. Not needed for
	Text *string `json:"text,omitempty"`
	// Text that would have been generated by the keyboard if no modifiers were pressed (except for
	UnmodifiedText *string `json:"unmodifiedText,omitempty"`
	// Unique key identifier (e.g., 'U+0041') (default: "").
	KeyIdentifier *string `json:"keyIdentifier,omitempty"`
	// Unique DOM defined string value for each physical key (e.g., 'KeyA') (default: "").
	Code *string `json:"code,omitempty"`
	// Unique DOM defined string value describing the meaning of the key in the context of active
	Key *string `json:"key,omitempty"`
	// Windows virtual key code (default: 0).
	WindowsVirtualKeyCode *int `json:"windowsVirtualKeyCode,omitempty"`
	// Native virtual key code (default: 0).
	NativeVirtualKeyCode *int `json:"nativeVirtualKeyCode,omitempty"`
	// Whether the event was generated from auto repeat (default: false).
	AutoRepeat *bool `json:"autoRepeat,omitempty"`
	// Whether the event was generated from the keypad (default: false).
	IsKeypad *bool `json:"isKeypad,omitempty"`
	// Whether the event was a system key event (default: false).
	IsSystemKey *bool `json:"isSystemKey,omitempty"`
	// Whether the event was from the left or right side of the keyboard. 1=Left, 2=Right (default:
	Location *int `json:"location,omitempty"`
	// Editing commands to send with the key event (e.g., 'selectAll') (default: []).
	Commands []string `json:"commands,omitempty"`
}

type InputDispatchKeyEventResult struct {
}

type InputInsertTextParams struct {
	SessionID string `json:"-"`
	// The text to insert.
	Text string `json:"text"`
}

type InputInsertTextResult struct {
}

type InputImeSetCompositionParams struct {
	SessionID string `json:"-"`
	// The text to insert
	Text string `json:"text"`
	// selection start
	SelectionStart int `json:"selectionStart"`
	// selection end
	SelectionEnd int `json:"selectionEnd"`
	// replacement start
	ReplacementStart *int `json:"replacementStart,omitempty"`
	// replacement end
	ReplacementEnd *int `json:"replacementEnd,omitempty"`
}

type InputImeSetCompositionResult struct {
}

type InputDispatchMouseEventParams struct {
	SessionID string `json:"-"`
	// Type of the mouse event.
	Type string `json:"type"`
	// X coordinate of the event relative to the main frame's viewport in CSS pixels.
	X float64 `json:"x"`
	// Y coordinate of the event relative to the main frame's viewport in CSS pixels. 0 refers to
	Y float64 `json:"y"`
	// Bit field representing pressed modifier keys. Alt=1, Ctrl=2, Meta/Command=4, Shift=8
	Modifiers *int `json:"modifiers,omitempty"`
	// Time at which the event occurred.
	Timestamp *InputTimeSinceEpoch `json:"timestamp,omitempty"`
	// Mouse button (default: "none").
	Button *InputMouseButton `json:"button,omitempty"`
	// A number indicating which buttons are pressed on the mouse when a mouse event is triggered.
	Buttons *int `json:"buttons,omitempty"`
	// Number of times the mouse button was clicked (default: 0).
	ClickCount *int `json:"clickCount,omitempty"`
	// The normalized pressure, which has a range of [0,1] (default: 0).
	Force *float64 `json:"force,omitempty"`
	// The normalized tangential pressure, which has a range of [-1,1] (default: 0).
	TangentialPressure *float64 `json:"tangentialPressure,omitempty"`
	// The plane angle between the Y-Z plane and the plane containing both the stylus axis and the Y axis, in degrees of the range [-90,90], a positive tiltX is to the right (default: 0).
	TiltX *float64 `json:"tiltX,omitempty"`
	// The plane angle between the X-Z plane and the plane containing both the stylus axis and the X axis, in degrees of the range [-90,90], a positive tiltY is towards the user (default:
	TiltY *float64 `json:"tiltY,omitempty"`
	// The clockwise rotation of a pen stylus around its own major axis, in degrees in the range [0,359] (default: 0).
	Twist *int `json:"twist,omitempty"`
	// X delta in CSS pixels for mouse wheel event (default: 0).
	DeltaX *float64 `json:"deltaX,omitempty"`
	// Y delta in CSS pixels for mouse wheel event (default: 0).
	DeltaY *float64 `json:"deltaY,omitempty"`
	// Pointer type (default: "mouse").
	PointerType *string `json:"pointerType,omitempty"`
}

type InputDispatchMouseEventResult struct {
}

type InputDispatchTouchEventParams struct {
	SessionID string `json:"-"`
	// Type of the touch event. TouchEnd and TouchCancel must not contain any touch points, while
	Type string `json:"type"`
	// Active touch points on the touch device. One event per any changed point (compared to
	TouchPoints []InputTouchPoint `json:"touchPoints"`
	// Bit field representing pressed modifier keys. Alt=1, Ctrl=2, Meta/Command=4, Shift=8
	Modifiers *int `json:"modifiers,omitempty"`
	// Time at which the event occurred.
	Timestamp *InputTimeSinceEpoch `json:"timestamp,omitempty"`
}

type InputDispatchTouchEventResult struct {
}

type InputCancelDraggingParams struct {
	SessionID string `json:"-"`
}

type InputCancelDraggingResult struct {
}

type InputEmulateTouchFromMouseEventParams struct {
	SessionID string `json:"-"`
	// Type of the mouse event.
	Type string `json:"type"`
	// X coordinate of the mouse pointer in DIP.
	X int `json:"x"`
	// Y coordinate of the mouse pointer in DIP.
	Y int `json:"y"`
	// Mouse button. Only "none", "left", "right" are supported.
	Button InputMouseButton `json:"button"`
	// Time at which the event occurred (default: current time).
	Timestamp *InputTimeSinceEpoch `json:"timestamp,omitempty"`
	// X delta in DIP for mouse wheel event (default: 0).
	DeltaX *float64 `json:"deltaX,omitempty"`
	// Y delta in DIP for mouse wheel event (default: 0).
	DeltaY *float64 `json:"deltaY,omitempty"`
	// Bit field representing pressed modifier keys. Alt=1, Ctrl=2, Meta/Command=4, Shift=8
	Modifiers *int `json:"modifiers,omitempty"`
	// Number of times the mouse button was clicked (default: 0).
	ClickCount *int `json:"clickCount,omitempty"`
}

type InputEmulateTouchFromMouseEventResult struct {
}

type InputSetIgnoreInputEventsParams struct {
	SessionID string `json:"-"`
	// Ignores input events processing when set to true.
	Ignore bool `json:"ignore"`
}

type InputSetIgnoreInputEventsResult struct {
}

type InputSetInterceptDragsParams struct {
	SessionID string `json:"-"`
	Enabled   bool   `json:"enabled"`
}

type InputSetInterceptDragsResult struct {
}

type InputSynthesizePinchGestureParams struct {
	SessionID string `json:"-"`
	// X coordinate of the start of the gesture in CSS pixels.
	X float64 `json:"x"`
	// Y coordinate of the start of the gesture in CSS pixels.
	Y float64 `json:"y"`
	// Relative scale factor after zooming (>1.0 zooms in, <1.0 zooms out).
	ScaleFactor float64 `json:"scaleFactor"`
	// Relative pointer speed in pixels per second (default: 800).
	RelativeSpeed *int `json:"relativeSpeed,omitempty"`
	// Which type of input events to be generated (default: 'default', which queries the platform
	GestureSourceType *InputGestureSourceType `json:"gestureSourceType,omitempty"`
}

type InputSynthesizePinchGestureResult struct {
}

type InputSynthesizeScrollGestureParams struct {
	SessionID string `json:"-"`
	// X coordinate of the start of the gesture in CSS pixels.
	X float64 `json:"x"`
	// Y coordinate of the start of the gesture in CSS pixels.
	Y float64 `json:"y"`
	// The distance to scroll along the X axis (positive to scroll left).
	XDistance *float64 `json:"xDistance,omitempty"`
	// The distance to scroll along the Y axis (positive to scroll up).
	YDistance *float64 `json:"yDistance,omitempty"`
	// The number of additional pixels to scroll back along the X axis, in addition to the given
	XOverscroll *float64 `json:"xOverscroll,omitempty"`
	// The number of additional pixels to scroll back along the Y axis, in addition to the given
	YOverscroll *float64 `json:"yOverscroll,omitempty"`
	// Prevent fling (default: true).
	PreventFling *bool `json:"preventFling,omitempty"`
	// Swipe speed in pixels per second (default: 800).
	Speed *int `json:"speed,omitempty"`
	// Which type of input events to be generated (default: 'default', which queries the platform
	GestureSourceType *InputGestureSourceType `json:"gestureSourceType,omitempty"`
	// The number of times to repeat the gesture (default: 0).
	RepeatCount *int `json:"repeatCount,omitempty"`
	// The number of milliseconds delay between each repeat. (default: 250).
	RepeatDelayMs *int `json:"repeatDelayMs,omitempty"`
	// The name of the interaction markers to generate, if not empty (default: "").
	InteractionMarkerName *string `json:"interactionMarkerName,omitempty"`
}

type InputSynthesizeScrollGestureResult struct {
}

type InputSynthesizeTapGestureParams struct {
	SessionID string `json:"-"`
	// X coordinate of the start of the gesture in CSS pixels.
	X float64 `json:"x"`
	// Y coordinate of the start of the gesture in CSS pixels.
	Y float64 `json:"y"`
	// Duration between touchdown and touchup events in ms (default: 50).
	Duration *int `json:"duration,omitempty"`
	// Number of times to perform the tap (e.g. 2 for double tap, default: 1).
	TapCount *int `json:"tapCount,omitempty"`
	// Which type of input events to be generated (default: 'default', which queries the platform
	GestureSourceType *InputGestureSourceType `json:"gestureSourceType,omitempty"`
}

type InputSynthesizeTapGestureResult struct {
}

type InputDragInterceptedEvent struct {
	Data InputDragData `json:"data"`
}

type InspectorDisableParams struct {
	SessionID string `json:"-"`
}

type InspectorDisableResult struct {
}

type InspectorEnableParams struct {
	SessionID string `json:"-"`
}

type InspectorEnableResult struct {
}

type InspectorDetachedEvent struct {
	// The reason why connection has been terminated.
	Reason string `json:"reason"`
}

type InspectorTargetCrashedEvent struct {
}

type InspectorTargetReloadedAfterCrashEvent struct {
}

type InspectorWorkerScriptLoadedEvent struct {
}

type LayerTreeLayerID string

type LayerTreeSnapshotID string

type LayerTreeScrollRect struct {
	// Rectangle itself.
	Rect DOMRect `json:"rect"`
	// Reason for rectangle to force scrolling on the main thread
	Type string `json:"type"`
}

type LayerTreeStickyPositionConstraint struct {
	// Layout rectangle of the sticky element before being shifted
	StickyBoxRect DOMRect `json:"stickyBoxRect"`
	// Layout rectangle of the containing block of the sticky element
	ContainingBlockRect DOMRect `json:"containingBlockRect"`
	// The nearest sticky layer that shifts the sticky box
	NearestLayerShiftingStickyBox *LayerTreeLayerID `json:"nearestLayerShiftingStickyBox,omitempty"`
	// The nearest sticky layer that shifts the containing block
	NearestLayerShiftingContainingBlock *LayerTreeLayerID `json:"nearestLayerShiftingContainingBlock,omitempty"`
}

type LayerTreePictureTile struct {
	// Offset from owning layer left boundary
	X float64 `json:"x"`
	// Offset from owning layer top boundary
	Y float64 `json:"y"`
	// Base64-encoded snapshot data. (Encoded as a base64 string when passed over JSON)
	Picture string `json:"picture"`
}

type LayerTreeLayer struct {
	// The unique id for this layer.
	LayerID LayerTreeLayerID `json:"layerId"`
	// The id of parent (not present for root).
	ParentLayerID *LayerTreeLayerID `json:"parentLayerId,omitempty"`
	// The backend id for the node associated with this layer.
	BackendNodeID *DOMBackendNodeID `json:"backendNodeId,omitempty"`
	// Offset from parent layer, X coordinate.
	OffsetX float64 `json:"offsetX"`
	// Offset from parent layer, Y coordinate.
	OffsetY float64 `json:"offsetY"`
	// Layer width.
	Width float64 `json:"width"`
	// Layer height.
	Height float64 `json:"height"`
	// Transformation matrix for layer, default is identity matrix
	Transform []float64 `json:"transform,omitempty"`
	// Transform anchor point X, absent if no transform specified
	AnchorX *float64 `json:"anchorX,omitempty"`
	// Transform anchor point Y, absent if no transform specified
	AnchorY *float64 `json:"anchorY,omitempty"`
	// Transform anchor point Z, absent if no transform specified
	AnchorZ *float64 `json:"anchorZ,omitempty"`
	// Indicates how many time this layer has painted.
	PaintCount int `json:"paintCount"`
	// Indicates whether this layer hosts any content, rather than being used for
	DrawsContent bool `json:"drawsContent"`
	// Set if layer is not visible.
	Invisible *bool `json:"invisible,omitempty"`
	// Rectangles scrolling on main thread only.
	ScrollRects []LayerTreeScrollRect `json:"scrollRects,omitempty"`
	// Sticky position constraint information
	StickyPositionConstraint *LayerTreeStickyPositionConstraint `json:"stickyPositionConstraint,omitempty"`
}

type LayerTreePaintProfile []float64

type LayerTreeCompositingReasonsParams struct {
	SessionID string `json:"-"`
	// The id of the layer for which we want to get the reasons it was composited.
	LayerID LayerTreeLayerID `json:"layerId"`
}

type LayerTreeCompositingReasonsResult struct {
	// A list of strings specifying reasons for the given layer to become composited.
	CompositingReasons []string `json:"compositingReasons"`
	// A list of strings specifying reason IDs for the given layer to become composited.
	CompositingReasonIds []string `json:"compositingReasonIds"`
}

type LayerTreeDisableParams struct {
	SessionID string `json:"-"`
}

type LayerTreeDisableResult struct {
}

type LayerTreeEnableParams struct {
	SessionID string `json:"-"`
}

type LayerTreeEnableResult struct {
}

type LayerTreeLoadSnapshotParams struct {
	SessionID string `json:"-"`
	// An array of tiles composing the snapshot.
	Tiles []LayerTreePictureTile `json:"tiles"`
}

type LayerTreeLoadSnapshotResult struct {
	// The id of the snapshot.
	SnapshotID LayerTreeSnapshotID `json:"snapshotId"`
}

type LayerTreeMakeSnapshotParams struct {
	SessionID string `json:"-"`
	// The id of the layer.
	LayerID LayerTreeLayerID `json:"layerId"`
}

type LayerTreeMakeSnapshotResult struct {
	// The id of the layer snapshot.
	SnapshotID LayerTreeSnapshotID `json:"snapshotId"`
}

type LayerTreeProfileSnapshotParams struct {
	SessionID string `json:"-"`
	// The id of the layer snapshot.
	SnapshotID LayerTreeSnapshotID `json:"snapshotId"`
	// The maximum number of times to replay the snapshot (1, if not specified).
	MinRepeatCount *int `json:"minRepeatCount,omitempty"`
	// The minimum duration (in seconds) to replay the snapshot.
	MinDuration *float64 `json:"minDuration,omitempty"`
	// The clip rectangle to apply when replaying the snapshot.
	ClipRect *DOMRect `json:"clipRect,omitempty"`
}

type LayerTreeProfileSnapshotResult struct {
	// The array of paint profiles, one per run.
	Timings []LayerTreePaintProfile `json:"timings"`
}

type LayerTreeReleaseSnapshotParams struct {
	SessionID string `json:"-"`
	// The id of the layer snapshot.
	SnapshotID LayerTreeSnapshotID `json:"snapshotId"`
}

type LayerTreeReleaseSnapshotResult struct {
}

type LayerTreeReplaySnapshotParams struct {
	SessionID string `json:"-"`
	// The id of the layer snapshot.
	SnapshotID LayerTreeSnapshotID `json:"snapshotId"`
	// The first step to replay from (replay from the very start if not specified).
	FromStep *int `json:"fromStep,omitempty"`
	// The last step to replay to (replay till the end if not specified).
	ToStep *int `json:"toStep,omitempty"`
	// The scale to apply while replaying (defaults to 1).
	Scale *float64 `json:"scale,omitempty"`
}

type LayerTreeReplaySnapshotResult struct {
	// A data: URL for resulting image.
	DataURL string `json:"dataURL"`
}

type LayerTreeSnapshotCommandLogParams struct {
	SessionID string `json:"-"`
	// The id of the layer snapshot.
	SnapshotID LayerTreeSnapshotID `json:"snapshotId"`
}

type LayerTreeSnapshotCommandLogResult struct {
	// The array of canvas function calls.
	CommandLog []map[string]any `json:"commandLog"`
}

type LayerTreeLayerPaintedEvent struct {
	// The id of the painted layer.
	LayerID LayerTreeLayerID `json:"layerId"`
	// Clip rectangle.
	Clip DOMRect `json:"clip"`
}

type LayerTreeLayerTreeDidChangeEvent struct {
	// Layer tree, absent if not in the compositing mode.
	Layers []LayerTreeLayer `json:"layers,omitempty"`
}

type LogLogEntry struct {
	// Log entry source.
	Source string `json:"source"`
	// Log entry severity.
	Level string `json:"level"`
	// Logged text.
	Text     string  `json:"text"`
	Category *string `json:"category,omitempty"`
	// Timestamp when this entry was added.
	Timestamp RuntimeTimestamp `json:"timestamp"`
	// URL of the resource if known.
	URL *string `json:"url,omitempty"`
	// Line number in the resource.
	LineNumber *int `json:"lineNumber,omitempty"`
	// JavaScript stack trace.
	StackTrace *RuntimeStackTrace `json:"stackTrace,omitempty"`
	// Identifier of the network request associated with this entry.
	NetworkRequestID *NetworkRequestID `json:"networkRequestId,omitempty"`
	// Identifier of the worker associated with this entry.
	WorkerID *string `json:"workerId,omitempty"`
	// Call arguments.
	Args []RuntimeRemoteObject `json:"args,omitempty"`
}

type LogViolationSetting struct {
	// Violation type.
	Name string `json:"name"`
	// Time threshold to trigger upon.
	Threshold float64 `json:"threshold"`
}

type LogClearParams struct {
	SessionID string `json:"-"`
}

type LogClearResult struct {
}

type LogDisableParams struct {
	SessionID string `json:"-"`
}

type LogDisableResult struct {
}

type LogEnableParams struct {
	SessionID string `json:"-"`
}

type LogEnableResult struct {
}

type LogStartViolationsReportParams struct {
	SessionID string `json:"-"`
	// Configuration for violations.
	Config []LogViolationSetting `json:"config"`
}

type LogStartViolationsReportResult struct {
}

type LogStopViolationsReportParams struct {
	SessionID string `json:"-"`
}

type LogStopViolationsReportResult struct {
}

type LogEntryAddedEvent struct {
	// The entry.
	Entry LogLogEntry `json:"entry"`
}

type MediaPlayerID string

type MediaTimestamp float64

type MediaPlayerMessage struct {
	// Keep in sync with MediaLogMessageLevel
	Level   string `json:"level"`
	Message string `json:"message"`
}

type MediaPlayerProperty struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type MediaPlayerEvent struct {
	Timestamp MediaTimestamp `json:"timestamp"`
	Value     string         `json:"value"`
}

type MediaPlayerErrorSourceLocation struct {
	File string `json:"file"`
	Line int    `json:"line"`
}

type MediaPlayerError struct {
	ErrorType string `json:"errorType"`
	// Code is the numeric enum entry for a specific set of error codes, such
	Code int `json:"code"`
	// A trace of where this error was caused / where it passed through.
	Stack []MediaPlayerErrorSourceLocation `json:"stack"`
	// Errors potentially have a root cause error, ie, a DecoderError might be
	Cause []MediaPlayerError `json:"cause"`
	// Extra data attached to an error, such as an HRESULT, Video Codec, etc.
	Data map[string]any `json:"data"`
}

type MediaPlayer struct {
	PlayerID  MediaPlayerID     `json:"playerId"`
	DOMNodeID *DOMBackendNodeID `json:"domNodeId,omitempty"`
}

type MediaEnableParams struct {
	SessionID string `json:"-"`
}

type MediaEnableResult struct {
}

type MediaDisableParams struct {
	SessionID string `json:"-"`
}

type MediaDisableResult struct {
}

type MediaPlayerPropertiesChangedEvent struct {
	PlayerID   MediaPlayerID         `json:"playerId"`
	Properties []MediaPlayerProperty `json:"properties"`
}

type MediaPlayerEventsAddedEvent struct {
	PlayerID MediaPlayerID      `json:"playerId"`
	Events   []MediaPlayerEvent `json:"events"`
}

type MediaPlayerMessagesLoggedEvent struct {
	PlayerID MediaPlayerID        `json:"playerId"`
	Messages []MediaPlayerMessage `json:"messages"`
}

type MediaPlayerErrorsRaisedEvent struct {
	PlayerID MediaPlayerID      `json:"playerId"`
	Errors   []MediaPlayerError `json:"errors"`
}

type MediaPlayerCreatedEvent struct {
	Player MediaPlayer `json:"player"`
}

type MemoryPressureLevel string

type MemorySamplingProfileNode struct {
	// Size of the sampled allocation.
	Size float64 `json:"size"`
	// Total bytes attributed to this sample.
	Total float64 `json:"total"`
	// Execution stack at the point of allocation.
	Stack []string `json:"stack"`
}

type MemorySamplingProfile struct {
	Samples []MemorySamplingProfileNode `json:"samples"`
	Modules []MemoryModule              `json:"modules"`
}

type MemoryModule struct {
	// Name of the module.
	Name string `json:"name"`
	// UUID of the module.
	UUID string `json:"uuid"`
	// Base address where the module is loaded into memory. Encoded as a decimal
	BaseAddress string `json:"baseAddress"`
	// Size of the module in bytes.
	Size float64 `json:"size"`
}

type MemoryDOMCounter struct {
	// Object name. Note: object names should be presumed volatile and clients should not expect
	Name string `json:"name"`
	// Object count.
	Count int `json:"count"`
}

type MemoryGetDOMCountersParams struct {
	SessionID string `json:"-"`
}

type MemoryGetDOMCountersResult struct {
	Documents        int `json:"documents"`
	Nodes            int `json:"nodes"`
	JSEventListeners int `json:"jsEventListeners"`
}

type MemoryGetDOMCountersForLeakDetectionParams struct {
	SessionID string `json:"-"`
}

type MemoryGetDOMCountersForLeakDetectionResult struct {
	// DOM object counters.
	Counters []MemoryDOMCounter `json:"counters"`
}

type MemoryPrepareForLeakDetectionParams struct {
	SessionID string `json:"-"`
}

type MemoryPrepareForLeakDetectionResult struct {
}

type MemoryForciblyPurgeJavaScriptMemoryParams struct {
	SessionID string `json:"-"`
}

type MemoryForciblyPurgeJavaScriptMemoryResult struct {
}

type MemorySetPressureNotificationsSuppressedParams struct {
	SessionID string `json:"-"`
	// If true, memory pressure notifications will be suppressed.
	Suppressed bool `json:"suppressed"`
}

type MemorySetPressureNotificationsSuppressedResult struct {
}

type MemorySimulatePressureNotificationParams struct {
	SessionID string `json:"-"`
	// Memory pressure level of the notification.
	Level MemoryPressureLevel `json:"level"`
}

type MemorySimulatePressureNotificationResult struct {
}

type MemoryStartSamplingParams struct {
	SessionID string `json:"-"`
	// Average number of bytes between samples.
	SamplingInterval *int `json:"samplingInterval,omitempty"`
	// Do not randomize intervals between samples.
	SuppressRandomness *bool `json:"suppressRandomness,omitempty"`
}

type MemoryStartSamplingResult struct {
}

type MemoryStopSamplingParams struct {
	SessionID string `json:"-"`
}

type MemoryStopSamplingResult struct {
}

type MemoryGetAllTimeSamplingProfileParams struct {
	SessionID string `json:"-"`
}

type MemoryGetAllTimeSamplingProfileResult struct {
	Profile MemorySamplingProfile `json:"profile"`
}

type MemoryGetBrowserSamplingProfileParams struct {
	SessionID string `json:"-"`
}

type MemoryGetBrowserSamplingProfileResult struct {
	Profile MemorySamplingProfile `json:"profile"`
}

type MemoryGetSamplingProfileParams struct {
	SessionID string `json:"-"`
}

type MemoryGetSamplingProfileResult struct {
	Profile MemorySamplingProfile `json:"profile"`
}

type NetworkResourceType string

type NetworkLoaderID string

type NetworkRequestID string

type NetworkInterceptionID string

type NetworkErrorReason string

type NetworkTimeSinceEpoch float64

type NetworkMonotonicTime float64

type NetworkHeaders map[string]any

type NetworkConnectionType string

type NetworkCookieSameSite string

type NetworkCookiePriority string

type NetworkCookieSourceScheme string

type NetworkResourceTiming struct {
	// Timing's requestTime is a baseline in seconds, while the other numbers are ticks in
	RequestTime float64 `json:"requestTime"`
	// Started resolving proxy.
	ProxyStart float64 `json:"proxyStart"`
	// Finished resolving proxy.
	ProxyEnd float64 `json:"proxyEnd"`
	// Started DNS address resolve.
	DnsStart float64 `json:"dnsStart"`
	// Finished DNS address resolve.
	DnsEnd float64 `json:"dnsEnd"`
	// Started connecting to the remote host.
	ConnectStart float64 `json:"connectStart"`
	// Connected to the remote host.
	ConnectEnd float64 `json:"connectEnd"`
	// Started SSL handshake.
	SslStart float64 `json:"sslStart"`
	// Finished SSL handshake.
	SslEnd float64 `json:"sslEnd"`
	// Started running ServiceWorker.
	WorkerStart float64 `json:"workerStart"`
	// Finished Starting ServiceWorker.
	WorkerReady float64 `json:"workerReady"`
	// Started fetch event.
	WorkerFetchStart float64 `json:"workerFetchStart"`
	// Settled fetch event respondWith promise.
	WorkerRespondWithSettled float64 `json:"workerRespondWithSettled"`
	// Started ServiceWorker static routing source evaluation.
	WorkerRouterEvaluationStart *float64 `json:"workerRouterEvaluationStart,omitempty"`
	// Started cache lookup when the source was evaluated to `cache`.
	WorkerCacheLookupStart *float64 `json:"workerCacheLookupStart,omitempty"`
	// Started sending request.
	SendStart float64 `json:"sendStart"`
	// Finished sending request.
	SendEnd float64 `json:"sendEnd"`
	// Time the server started pushing request.
	PushStart float64 `json:"pushStart"`
	// Time the server finished pushing request.
	PushEnd float64 `json:"pushEnd"`
	// Started receiving response headers.
	ReceiveHeadersStart float64 `json:"receiveHeadersStart"`
	// Finished receiving response headers.
	ReceiveHeadersEnd float64 `json:"receiveHeadersEnd"`
}

type NetworkResourcePriority string

type NetworkRenderBlockingBehavior string

type NetworkPostDataEntry struct {
	Bytes *string `json:"bytes,omitempty"`
}

type NetworkRequest struct {
	// Request URL (without fragment).
	URL string `json:"url"`
	// Fragment of the requested URL starting with hash, if present.
	URLFragment *string `json:"urlFragment,omitempty"`
	// HTTP request method.
	Method string `json:"method"`
	// HTTP request headers.
	Headers NetworkHeaders `json:"headers"`
	// HTTP POST request data.
	PostData *string `json:"postData,omitempty"`
	// True when the request has POST data. Note that postData might still be omitted when this flag is true when the data is too long.
	HasPostData *bool `json:"hasPostData,omitempty"`
	// Request body elements (post data broken into individual entries).
	PostDataEntries []NetworkPostDataEntry `json:"postDataEntries,omitempty"`
	// The mixed content type of the request.
	MixedContentType *SecurityMixedContentType `json:"mixedContentType,omitempty"`
	// Priority of the resource request at the time request is sent.
	InitialPriority NetworkResourcePriority `json:"initialPriority"`
	// The referrer policy of the request, as defined in https://www.w3.org/TR/referrer-policy/
	ReferrerPolicy string `json:"referrerPolicy"`
	// Whether is loaded via link preload.
	IsLinkPreload *bool `json:"isLinkPreload,omitempty"`
	// Set for requests when the TrustToken API is used. Contains the parameters
	TrustTokenParams *NetworkTrustTokenParams `json:"trustTokenParams,omitempty"`
	// True if this resource request is considered to be the 'same site' as the
	IsSameSite *bool `json:"isSameSite,omitempty"`
	// True when the resource request is ad-related.
	IsAdRelated *bool `json:"isAdRelated,omitempty"`
}

type NetworkSignedCertificateTimestamp struct {
	// Validation status.
	Status string `json:"status"`
	// Origin.
	Origin string `json:"origin"`
	// Log name / description.
	LogDescription string `json:"logDescription"`
	// Log ID.
	LogID string `json:"logId"`
	// Issuance date. Unlike TimeSinceEpoch, this contains the number of
	Timestamp float64 `json:"timestamp"`
	// Hash algorithm.
	HashAlgorithm string `json:"hashAlgorithm"`
	// Signature algorithm.
	SignatureAlgorithm string `json:"signatureAlgorithm"`
	// Signature data.
	SignatureData string `json:"signatureData"`
}

type NetworkSecurityDetails struct {
	// Protocol name (e.g. "TLS 1.2" or "QUIC").
	Protocol string `json:"protocol"`
	// Key Exchange used by the connection, or the empty string if not applicable.
	KeyExchange string `json:"keyExchange"`
	// (EC)DH group used by the connection, if applicable.
	KeyExchangeGroup *string `json:"keyExchangeGroup,omitempty"`
	// Cipher name.
	Cipher string `json:"cipher"`
	// TLS MAC. Note that AEAD ciphers do not have separate MACs.
	Mac *string `json:"mac,omitempty"`
	// Certificate ID value.
	CertificateID SecurityCertificateID `json:"certificateId"`
	// Certificate subject name.
	SubjectName string `json:"subjectName"`
	// Subject Alternative Name (SAN) DNS names and IP addresses.
	SanList []string `json:"sanList"`
	// Name of the issuing CA.
	Issuer string `json:"issuer"`
	// Certificate valid from date.
	ValidFrom NetworkTimeSinceEpoch `json:"validFrom"`
	// Certificate valid to (expiration) date
	ValidTo NetworkTimeSinceEpoch `json:"validTo"`
	// List of signed certificate timestamps (SCTs).
	SignedCertificateTimestampList []NetworkSignedCertificateTimestamp `json:"signedCertificateTimestampList"`
	// Whether the request complied with Certificate Transparency policy
	CertificateTransparencyCompliance NetworkCertificateTransparencyCompliance `json:"certificateTransparencyCompliance"`
	// The signature algorithm used by the server in the TLS server signature,
	ServerSignatureAlgorithm *int `json:"serverSignatureAlgorithm,omitempty"`
	// Whether the connection used Encrypted ClientHello
	EncryptedClientHello bool `json:"encryptedClientHello"`
}

type NetworkCertificateTransparencyCompliance string

type NetworkBlockedReason string

type NetworkCorsError string

type NetworkCorsErrorStatus struct {
	CorsError       NetworkCorsError `json:"corsError"`
	FailedParameter string           `json:"failedParameter"`
}

type NetworkServiceWorkerResponseSource string

type NetworkTrustTokenParams struct {
	Operation NetworkTrustTokenOperationType `json:"operation"`
	// Only set for "token-redemption" operation and determine whether
	RefreshPolicy string `json:"refreshPolicy"`
	// Origins of issuers from whom to request tokens or redemption
	Issuers []string `json:"issuers,omitempty"`
}

type NetworkTrustTokenOperationType string

type NetworkAlternateProtocolUsage string

type NetworkServiceWorkerRouterSource string

type NetworkServiceWorkerRouterInfo struct {
	// ID of the rule matched. If there is a matched rule, this field will
	RuleIDMatched *int `json:"ruleIdMatched,omitempty"`
	// The router source of the matched rule. If there is a matched rule, this
	MatchedSourceType *NetworkServiceWorkerRouterSource `json:"matchedSourceType,omitempty"`
	// The actual router source used.
	ActualSourceType *NetworkServiceWorkerRouterSource `json:"actualSourceType,omitempty"`
}

type NetworkResponse struct {
	// Response URL. This URL can be different from CachedResource.url in case of redirect.
	URL string `json:"url"`
	// HTTP response status code.
	Status int `json:"status"`
	// HTTP response status text.
	StatusText string `json:"statusText"`
	// HTTP response headers.
	Headers NetworkHeaders `json:"headers"`
	// HTTP response headers text. This has been replaced by the headers in Network.responseReceivedExtraInfo.
	HeadersText *string `json:"headersText,omitempty"`
	// Resource mimeType as determined by the browser.
	MimeType string `json:"mimeType"`
	// Resource charset as determined by the browser (if applicable).
	Charset string `json:"charset"`
	// Refined HTTP request headers that were actually transmitted over the network.
	RequestHeaders *NetworkHeaders `json:"requestHeaders,omitempty"`
	// HTTP request headers text. This has been replaced by the headers in Network.requestWillBeSentExtraInfo.
	RequestHeadersText *string `json:"requestHeadersText,omitempty"`
	// Specifies whether physical connection was actually reused for this request.
	ConnectionReused bool `json:"connectionReused"`
	// Physical connection id that was actually used for this request.
	ConnectionID float64 `json:"connectionId"`
	// Remote IP address.
	RemoteIPAddress *string `json:"remoteIPAddress,omitempty"`
	// Remote port.
	RemotePort *int `json:"remotePort,omitempty"`
	// Specifies that the request was served from the disk cache.
	FromDiskCache *bool `json:"fromDiskCache,omitempty"`
	// Specifies that the request was served from the ServiceWorker.
	FromServiceWorker *bool `json:"fromServiceWorker,omitempty"`
	// Specifies that the request was served from the prefetch cache.
	FromPrefetchCache *bool `json:"fromPrefetchCache,omitempty"`
	// Specifies that the request was served from the prefetch cache.
	FromEarlyHints *bool `json:"fromEarlyHints,omitempty"`
	// Information about how ServiceWorker Static Router API was used. If this
	ServiceWorkerRouterInfo *NetworkServiceWorkerRouterInfo `json:"serviceWorkerRouterInfo,omitempty"`
	// Total number of bytes received for this request so far.
	EncodedDataLength float64 `json:"encodedDataLength"`
	// Timing information for the given request.
	Timing *NetworkResourceTiming `json:"timing,omitempty"`
	// Response source of response from ServiceWorker.
	ServiceWorkerResponseSource *NetworkServiceWorkerResponseSource `json:"serviceWorkerResponseSource,omitempty"`
	// The time at which the returned response was generated.
	ResponseTime *NetworkTimeSinceEpoch `json:"responseTime,omitempty"`
	// Cache Storage Cache Name.
	CacheStorageCacheName *string `json:"cacheStorageCacheName,omitempty"`
	// Protocol used to fetch this request.
	Protocol *string `json:"protocol,omitempty"`
	// The reason why Chrome uses a specific transport protocol for HTTP semantics.
	AlternateProtocolUsage *NetworkAlternateProtocolUsage `json:"alternateProtocolUsage,omitempty"`
	// Security state of the request resource.
	SecurityState SecuritySecurityState `json:"securityState"`
	// Security details for the request.
	SecurityDetails *NetworkSecurityDetails `json:"securityDetails,omitempty"`
}

type NetworkWebSocketRequest struct {
	// HTTP request headers.
	Headers NetworkHeaders `json:"headers"`
}

type NetworkWebSocketResponse struct {
	// HTTP response status code.
	Status int `json:"status"`
	// HTTP response status text.
	StatusText string `json:"statusText"`
	// HTTP response headers.
	Headers NetworkHeaders `json:"headers"`
	// HTTP response headers text.
	HeadersText *string `json:"headersText,omitempty"`
	// HTTP request headers.
	RequestHeaders *NetworkHeaders `json:"requestHeaders,omitempty"`
	// HTTP request headers text.
	RequestHeadersText *string `json:"requestHeadersText,omitempty"`
}

type NetworkWebSocketFrame struct {
	// WebSocket message opcode.
	Opcode float64 `json:"opcode"`
	// WebSocket message mask.
	Mask bool `json:"mask"`
	// WebSocket message payload data.
	PayloadData string `json:"payloadData"`
}

type NetworkCachedResource struct {
	// Resource URL. This is the url of the original network request.
	URL string `json:"url"`
	// Type of this resource.
	Type NetworkResourceType `json:"type"`
	// Cached response data.
	Response *NetworkResponse `json:"response,omitempty"`
	// Cached response body size.
	BodySize float64 `json:"bodySize"`
}

type NetworkInitiator struct {
	// Type of this initiator.
	Type string `json:"type"`
	// Initiator JavaScript stack trace, set for Script only.
	Stack *RuntimeStackTrace `json:"stack,omitempty"`
	// Initiator URL, set for Parser type or for Script type (when script is importing module) or for SignedExchange type.
	URL *string `json:"url,omitempty"`
	// Initiator line number, set for Parser type or for Script type (when script is importing
	LineNumber *float64 `json:"lineNumber,omitempty"`
	// Initiator column number, set for Parser type or for Script type (when script is importing
	ColumnNumber *float64 `json:"columnNumber,omitempty"`
	// Set if another request triggered this request (e.g. preflight).
	RequestID *NetworkRequestID `json:"requestId,omitempty"`
}

type NetworkCookiePartitionKey struct {
	// The site of the top-level URL the browser was visiting at the start
	TopLevelSite string `json:"topLevelSite"`
	// Indicates if the cookie has any ancestors that are cross-site to the topLevelSite.
	HasCrossSiteAncestor bool `json:"hasCrossSiteAncestor"`
}

type NetworkCookie struct {
	// Cookie name.
	Name string `json:"name"`
	// Cookie value.
	Value string `json:"value"`
	// Cookie domain.
	Domain string `json:"domain"`
	// Cookie path.
	Path string `json:"path"`
	// Cookie expiration date as the number of seconds since the UNIX epoch.
	Expires float64 `json:"expires"`
	// Cookie size.
	Size int `json:"size"`
	// True if cookie is http-only.
	HTTPOnly bool `json:"httpOnly"`
	// True if cookie is secure.
	Secure bool `json:"secure"`
	// True in case of session cookie.
	Session bool `json:"session"`
	// Cookie SameSite type.
	SameSite *NetworkCookieSameSite `json:"sameSite,omitempty"`
	// Cookie Priority
	Priority NetworkCookiePriority `json:"priority"`
	// Cookie source scheme type.
	SourceScheme NetworkCookieSourceScheme `json:"sourceScheme"`
	// Cookie source port. Valid values are {-1, [1, 65535]}, -1 indicates an unspecified port.
	SourcePort int `json:"sourcePort"`
	// Cookie partition key.
	PartitionKey *NetworkCookiePartitionKey `json:"partitionKey,omitempty"`
	// True if cookie partition key is opaque.
	PartitionKeyOpaque *bool `json:"partitionKeyOpaque,omitempty"`
}

type NetworkSetCookieBlockedReason string

type NetworkCookieBlockedReason string

type NetworkCookieExemptionReason string

type NetworkBlockedSetCookieWithReason struct {
	// The reason(s) this cookie was blocked.
	BlockedReasons []NetworkSetCookieBlockedReason `json:"blockedReasons"`
	// The string representing this individual cookie as it would appear in the header.
	CookieLine string `json:"cookieLine"`
	// The cookie object which represents the cookie which was not stored. It is optional because
	Cookie *NetworkCookie `json:"cookie,omitempty"`
}

type NetworkExemptedSetCookieWithReason struct {
	// The reason the cookie was exempted.
	ExemptionReason NetworkCookieExemptionReason `json:"exemptionReason"`
	// The string representing this individual cookie as it would appear in the header.
	CookieLine string `json:"cookieLine"`
	// The cookie object representing the cookie.
	Cookie NetworkCookie `json:"cookie"`
}

type NetworkAssociatedCookie struct {
	// The cookie object representing the cookie which was not sent.
	Cookie NetworkCookie `json:"cookie"`
	// The reason(s) the cookie was blocked. If empty means the cookie is included.
	BlockedReasons []NetworkCookieBlockedReason `json:"blockedReasons"`
	// The reason the cookie should have been blocked by 3PCD but is exempted. A cookie could
	ExemptionReason *NetworkCookieExemptionReason `json:"exemptionReason,omitempty"`
}

type NetworkCookieParam struct {
	// Cookie name.
	Name string `json:"name"`
	// Cookie value.
	Value string `json:"value"`
	// The request-URI to associate with the setting of the cookie. This value can affect the
	URL *string `json:"url,omitempty"`
	// Cookie domain.
	Domain *string `json:"domain,omitempty"`
	// Cookie path.
	Path *string `json:"path,omitempty"`
	// True if cookie is secure.
	Secure *bool `json:"secure,omitempty"`
	// True if cookie is http-only.
	HTTPOnly *bool `json:"httpOnly,omitempty"`
	// Cookie SameSite type.
	SameSite *NetworkCookieSameSite `json:"sameSite,omitempty"`
	// Cookie expiration date, session cookie if not set
	Expires *NetworkTimeSinceEpoch `json:"expires,omitempty"`
	// Cookie Priority.
	Priority *NetworkCookiePriority `json:"priority,omitempty"`
	// Cookie source scheme type.
	SourceScheme *NetworkCookieSourceScheme `json:"sourceScheme,omitempty"`
	// Cookie source port. Valid values are {-1, [1, 65535]}, -1 indicates an unspecified port.
	SourcePort *int `json:"sourcePort,omitempty"`
	// Cookie partition key. If not set, the cookie will be set as not partitioned.
	PartitionKey *NetworkCookiePartitionKey `json:"partitionKey,omitempty"`
}

type NetworkAuthChallenge struct {
	// Source of the authentication challenge.
	Source *string `json:"source,omitempty"`
	// Origin of the challenger.
	Origin string `json:"origin"`
	// The authentication scheme used, such as basic or digest
	Scheme string `json:"scheme"`
	// The realm of the challenge. May be empty.
	Realm string `json:"realm"`
}

type NetworkAuthChallengeResponse struct {
	// The decision on what to do in response to the authorization challenge.  Default means
	Response string `json:"response"`
	// The username to provide, possibly empty. Should only be set if response is
	Username *string `json:"username,omitempty"`
	// The password to provide, possibly empty. Should only be set if response is
	Password *string `json:"password,omitempty"`
}

type NetworkInterceptionStage string

type NetworkRequestPattern struct {
	// Wildcards (`'*'` -> zero or more, `'?'` -> exactly one) are allowed. Escape character is
	URLPattern *string `json:"urlPattern,omitempty"`
	// If set, only requests for matching resource types will be intercepted.
	ResourceType *NetworkResourceType `json:"resourceType,omitempty"`
	// Stage at which to begin intercepting requests. Default is Request.
	InterceptionStage *NetworkInterceptionStage `json:"interceptionStage,omitempty"`
}

type NetworkSignedExchangeSignature struct {
	// Signed exchange signature label.
	Label string `json:"label"`
	// The hex string of signed exchange signature.
	Signature string `json:"signature"`
	// Signed exchange signature integrity.
	Integrity string `json:"integrity"`
	// Signed exchange signature cert Url.
	CertURL *string `json:"certUrl,omitempty"`
	// The hex string of signed exchange signature cert sha256.
	CertSha256 *string `json:"certSha256,omitempty"`
	// Signed exchange signature validity Url.
	ValidityURL string `json:"validityUrl"`
	// Signed exchange signature date.
	Date int `json:"date"`
	// Signed exchange signature expires.
	Expires int `json:"expires"`
	// The encoded certificates.
	Certificates []string `json:"certificates,omitempty"`
}

type NetworkSignedExchangeHeader struct {
	// Signed exchange request URL.
	RequestURL string `json:"requestUrl"`
	// Signed exchange response code.
	ResponseCode int `json:"responseCode"`
	// Signed exchange response headers.
	ResponseHeaders NetworkHeaders `json:"responseHeaders"`
	// Signed exchange response signature.
	Signatures []NetworkSignedExchangeSignature `json:"signatures"`
	// Signed exchange header integrity hash in the form of `sha256-<base64-hash-value>`.
	HeaderIntegrity string `json:"headerIntegrity"`
}

type NetworkSignedExchangeErrorField string

type NetworkSignedExchangeError struct {
	// Error message.
	Message string `json:"message"`
	// The index of the signature which caused the error.
	SignatureIndex *int `json:"signatureIndex,omitempty"`
	// The field which caused the error.
	ErrorField *NetworkSignedExchangeErrorField `json:"errorField,omitempty"`
}

type NetworkSignedExchangeInfo struct {
	// The outer response of signed HTTP exchange which was received from network.
	OuterResponse NetworkResponse `json:"outerResponse"`
	// Whether network response for the signed exchange was accompanied by
	HasExtraInfo bool `json:"hasExtraInfo"`
	// Information about the signed exchange header.
	Header *NetworkSignedExchangeHeader `json:"header,omitempty"`
	// Security details for the signed exchange header.
	SecurityDetails *NetworkSecurityDetails `json:"securityDetails,omitempty"`
	// Errors occurred while handling the signed exchange.
	Errors []NetworkSignedExchangeError `json:"errors,omitempty"`
}

type NetworkContentEncoding string

type NetworkNetworkConditions struct {
	// Only matching requests will be affected by these conditions. Patterns use the URLPattern constructor string
	URLPattern string `json:"urlPattern"`
	// Minimum latency from request sent to response headers received (ms).
	Latency float64 `json:"latency"`
	// Maximal aggregated download throughput (bytes/sec). -1 disables download throttling.
	DownloadThroughput float64 `json:"downloadThroughput"`
	// Maximal aggregated upload throughput (bytes/sec).  -1 disables upload throttling.
	UploadThroughput float64 `json:"uploadThroughput"`
	// Connection type if known.
	ConnectionType *NetworkConnectionType `json:"connectionType,omitempty"`
	// WebRTC packet loss (percent, 0-100). 0 disables packet loss emulation, 100 drops all the packets.
	PacketLoss *float64 `json:"packetLoss,omitempty"`
	// WebRTC packet queue length (packet). 0 removes any queue length limitations.
	PacketQueueLength *int `json:"packetQueueLength,omitempty"`
	// WebRTC packetReordering feature.
	PacketReordering *bool `json:"packetReordering,omitempty"`
}

type NetworkBlockPattern struct {
	// URL pattern to match. Patterns use the URLPattern constructor string syntax
	URLPattern string `json:"urlPattern"`
	// Whether or not to block the pattern. If false, a matching request will not be blocked even if it matches a later
	Block bool `json:"block"`
}

type NetworkDirectSocketDnsQueryType string

type NetworkDirectTCPSocketOptions struct {
	// TCP_NODELAY option
	NoDelay bool `json:"noDelay"`
	// Expected to be unsigned integer.
	KeepAliveDelay *float64 `json:"keepAliveDelay,omitempty"`
	// Expected to be unsigned integer.
	SendBufferSize *float64 `json:"sendBufferSize,omitempty"`
	// Expected to be unsigned integer.
	ReceiveBufferSize *float64                         `json:"receiveBufferSize,omitempty"`
	DnsQueryType      *NetworkDirectSocketDnsQueryType `json:"dnsQueryType,omitempty"`
}

type NetworkDirectUDPSocketOptions struct {
	RemoteAddr *string `json:"remoteAddr,omitempty"`
	// Unsigned int 16.
	RemotePort *int    `json:"remotePort,omitempty"`
	LocalAddr  *string `json:"localAddr,omitempty"`
	// Unsigned int 16.
	LocalPort    *int                             `json:"localPort,omitempty"`
	DnsQueryType *NetworkDirectSocketDnsQueryType `json:"dnsQueryType,omitempty"`
	// Expected to be unsigned integer.
	SendBufferSize *float64 `json:"sendBufferSize,omitempty"`
	// Expected to be unsigned integer.
	ReceiveBufferSize *float64 `json:"receiveBufferSize,omitempty"`
	MulticastLoopback *bool    `json:"multicastLoopback,omitempty"`
	// Unsigned int 8.
	MulticastTimeToLive          *int  `json:"multicastTimeToLive,omitempty"`
	MulticastAllowAddressSharing *bool `json:"multicastAllowAddressSharing,omitempty"`
}

type NetworkDirectUDPMessage struct {
	Data string `json:"data"`
	// Null for connected mode.
	RemoteAddr *string `json:"remoteAddr,omitempty"`
	// Null for connected mode.
	RemotePort *int `json:"remotePort,omitempty"`
}

type NetworkLocalNetworkAccessRequestPolicy string

type NetworkIPAddressSpace string

type NetworkConnectTiming struct {
	// Timing's requestTime is a baseline in seconds, while the other numbers are ticks in
	RequestTime float64 `json:"requestTime"`
}

type NetworkClientSecurityState struct {
	InitiatorIsSecureContext        bool                                   `json:"initiatorIsSecureContext"`
	InitiatorIPAddressSpace         NetworkIPAddressSpace                  `json:"initiatorIPAddressSpace"`
	LocalNetworkAccessRequestPolicy NetworkLocalNetworkAccessRequestPolicy `json:"localNetworkAccessRequestPolicy"`
}

type NetworkCrossOriginOpenerPolicyValue string

type NetworkCrossOriginOpenerPolicyStatus struct {
	Value                       NetworkCrossOriginOpenerPolicyValue `json:"value"`
	ReportOnlyValue             NetworkCrossOriginOpenerPolicyValue `json:"reportOnlyValue"`
	ReportingEndpoint           *string                             `json:"reportingEndpoint,omitempty"`
	ReportOnlyReportingEndpoint *string                             `json:"reportOnlyReportingEndpoint,omitempty"`
}

type NetworkCrossOriginEmbedderPolicyValue string

type NetworkCrossOriginEmbedderPolicyStatus struct {
	Value                       NetworkCrossOriginEmbedderPolicyValue `json:"value"`
	ReportOnlyValue             NetworkCrossOriginEmbedderPolicyValue `json:"reportOnlyValue"`
	ReportingEndpoint           *string                               `json:"reportingEndpoint,omitempty"`
	ReportOnlyReportingEndpoint *string                               `json:"reportOnlyReportingEndpoint,omitempty"`
}

type NetworkContentSecurityPolicySource string

type NetworkContentSecurityPolicyStatus struct {
	EffectiveDirectives string                             `json:"effectiveDirectives"`
	IsEnforced          bool                               `json:"isEnforced"`
	Source              NetworkContentSecurityPolicySource `json:"source"`
}

type NetworkSecurityIsolationStatus struct {
	Coop *NetworkCrossOriginOpenerPolicyStatus   `json:"coop,omitempty"`
	Coep *NetworkCrossOriginEmbedderPolicyStatus `json:"coep,omitempty"`
	Csp  []NetworkContentSecurityPolicyStatus    `json:"csp,omitempty"`
}

type NetworkReportStatus string

type NetworkReportID string

type NetworkReportingAPIReport struct {
	ID NetworkReportID `json:"id"`
	// The URL of the document that triggered the report.
	InitiatorURL string `json:"initiatorUrl"`
	// The name of the endpoint group that should be used to deliver the report.
	Destination string `json:"destination"`
	// The type of the report (specifies the set of data that is contained in the report body).
	Type string `json:"type"`
	// When the report was generated.
	Timestamp NetworkTimeSinceEpoch `json:"timestamp"`
	// How many uploads deep the related request was.
	Depth int `json:"depth"`
	// The number of delivery attempts made so far, not including an active attempt.
	CompletedAttempts int                 `json:"completedAttempts"`
	Body              map[string]any      `json:"body"`
	Status            NetworkReportStatus `json:"status"`
}

type NetworkReportingAPIEndpoint struct {
	// The URL of the endpoint to which reports may be delivered.
	URL string `json:"url"`
	// Name of the endpoint group.
	GroupName string `json:"groupName"`
}

type NetworkDeviceBoundSessionKey struct {
	// The site the session is set up for.
	Site string `json:"site"`
	// The id of the session.
	ID string `json:"id"`
}

type NetworkDeviceBoundSessionWithUsage struct {
	// The key for the session.
	SessionKey NetworkDeviceBoundSessionKey `json:"sessionKey"`
	// How the session was used (or not used).
	Usage string `json:"usage"`
}

type NetworkDeviceBoundSessionCookieCraving struct {
	// The name of the craving.
	Name string `json:"name"`
	// The domain of the craving.
	Domain string `json:"domain"`
	// The path of the craving.
	Path string `json:"path"`
	// The `Secure` attribute of the craving attributes.
	Secure bool `json:"secure"`
	// The `HttpOnly` attribute of the craving attributes.
	HTTPOnly bool `json:"httpOnly"`
	// The `SameSite` attribute of the craving attributes.
	SameSite *NetworkCookieSameSite `json:"sameSite,omitempty"`
}

type NetworkDeviceBoundSessionURLRule struct {
	// See comments on `net::device_bound_sessions::SessionInclusionRules::UrlRule::rule_type`.
	RuleType string `json:"ruleType"`
	// See comments on `net::device_bound_sessions::SessionInclusionRules::UrlRule::host_pattern`.
	HostPattern string `json:"hostPattern"`
	// See comments on `net::device_bound_sessions::SessionInclusionRules::UrlRule::path_prefix`.
	PathPrefix string `json:"pathPrefix"`
}

type NetworkDeviceBoundSessionInclusionRules struct {
	// See comments on `net::device_bound_sessions::SessionInclusionRules::origin_`.
	Origin string `json:"origin"`
	// Whether the whole site is included. See comments on
	IncludeSite bool `json:"includeSite"`
	// See comments on `net::device_bound_sessions::SessionInclusionRules::url_rules_`.
	URLRules []NetworkDeviceBoundSessionURLRule `json:"urlRules"`
}

type NetworkDeviceBoundSession struct {
	// The site and session ID of the session.
	Key NetworkDeviceBoundSessionKey `json:"key"`
	// See comments on `net::device_bound_sessions::Session::refresh_url_`.
	RefreshURL string `json:"refreshUrl"`
	// See comments on `net::device_bound_sessions::Session::inclusion_rules_`.
	InclusionRules NetworkDeviceBoundSessionInclusionRules `json:"inclusionRules"`
	// See comments on `net::device_bound_sessions::Session::cookie_cravings_`.
	CookieCravings []NetworkDeviceBoundSessionCookieCraving `json:"cookieCravings"`
	// See comments on `net::device_bound_sessions::Session::expiry_date_`.
	ExpiryDate NetworkTimeSinceEpoch `json:"expiryDate"`
	// See comments on `net::device_bound_sessions::Session::cached_challenge__`.
	CachedChallenge *string `json:"cachedChallenge,omitempty"`
	// See comments on `net::device_bound_sessions::Session::allowed_refresh_initiators_`.
	AllowedRefreshInitiators []string `json:"allowedRefreshInitiators"`
}

type NetworkDeviceBoundSessionEventID string

type NetworkDeviceBoundSessionFetchResult string

type NetworkDeviceBoundSessionFailedRequest struct {
	// The failed request URL.
	RequestURL string `json:"requestUrl"`
	// The net error of the response if it was not OK.
	NetError *string `json:"netError,omitempty"`
	// The response code if the net error was OK and the response code was not
	ResponseError *int `json:"responseError,omitempty"`
	// The body of the response if the net error was OK, the response code was
	ResponseErrorBody *string `json:"responseErrorBody,omitempty"`
}

type NetworkCreationEventDetails struct {
	// The result of the fetch attempt.
	FetchResult NetworkDeviceBoundSessionFetchResult `json:"fetchResult"`
	// The session if there was a newly created session. This is populated for
	NewSession *NetworkDeviceBoundSession `json:"newSession,omitempty"`
	// Details about a failed device bound session network request if there was
	FailedRequest *NetworkDeviceBoundSessionFailedRequest `json:"failedRequest,omitempty"`
}

type NetworkRefreshEventDetails struct {
	// The result of a refresh.
	RefreshResult string `json:"refreshResult"`
	// If there was a fetch attempt, the result of that.
	FetchResult *NetworkDeviceBoundSessionFetchResult `json:"fetchResult,omitempty"`
	// The session display if there was a newly created session. This is populated
	NewSession *NetworkDeviceBoundSession `json:"newSession,omitempty"`
	// See comments on `net::device_bound_sessions::RefreshEventResult::was_fully_proactive_refresh`.
	WasFullyProactiveRefresh bool `json:"wasFullyProactiveRefresh"`
	// Details about a failed device bound session network request if there was
	FailedRequest *NetworkDeviceBoundSessionFailedRequest `json:"failedRequest,omitempty"`
}

type NetworkTerminationEventDetails struct {
	// The reason for a session being deleted.
	DeletionReason string `json:"deletionReason"`
}

type NetworkChallengeEventDetails struct {
	// The result of a challenge.
	ChallengeResult string `json:"challengeResult"`
	// The challenge set.
	Challenge string `json:"challenge"`
}

type NetworkLoadNetworkResourcePageResult struct {
	Success bool `json:"success"`
	// Optional values used for error reporting.
	NetError       *float64 `json:"netError,omitempty"`
	NetErrorName   *string  `json:"netErrorName,omitempty"`
	HTTPStatusCode *float64 `json:"httpStatusCode,omitempty"`
	// If successful, one of the following two fields holds the result.
	Stream *IOStreamHandle `json:"stream,omitempty"`
	// Response headers.
	Headers *NetworkHeaders `json:"headers,omitempty"`
}

type NetworkLoadNetworkResourceOptions struct {
	DisableCache       bool `json:"disableCache"`
	IncludeCredentials bool `json:"includeCredentials"`
}

type NetworkSetAcceptedEncodingsParams struct {
	SessionID string `json:"-"`
	// List of accepted content encodings.
	Encodings []NetworkContentEncoding `json:"encodings"`
}

type NetworkSetAcceptedEncodingsResult struct {
}

type NetworkClearAcceptedEncodingsOverrideParams struct {
	SessionID string `json:"-"`
}

type NetworkClearAcceptedEncodingsOverrideResult struct {
}

type NetworkCanClearBrowserCacheParams struct {
	SessionID string `json:"-"`
}

type NetworkCanClearBrowserCacheResult struct {
	// True if browser cache can be cleared.
	Result bool `json:"result"`
}

type NetworkCanClearBrowserCookiesParams struct {
	SessionID string `json:"-"`
}

type NetworkCanClearBrowserCookiesResult struct {
	// True if browser cookies can be cleared.
	Result bool `json:"result"`
}

type NetworkCanEmulateNetworkConditionsParams struct {
	SessionID string `json:"-"`
}

type NetworkCanEmulateNetworkConditionsResult struct {
	// True if emulation of network conditions is supported.
	Result bool `json:"result"`
}

type NetworkClearBrowserCacheParams struct {
	SessionID string `json:"-"`
}

type NetworkClearBrowserCacheResult struct {
}

type NetworkClearBrowserCookiesParams struct {
	SessionID string `json:"-"`
}

type NetworkClearBrowserCookiesResult struct {
}

type NetworkContinueInterceptedRequestParams struct {
	SessionID      string                `json:"-"`
	InterceptionID NetworkInterceptionID `json:"interceptionId"`
	// If set this causes the request to fail with the given reason. Passing `Aborted` for requests
	ErrorReason *NetworkErrorReason `json:"errorReason,omitempty"`
	// If set the requests completes using with the provided base64 encoded raw response, including
	RawResponse *string `json:"rawResponse,omitempty"`
	// If set the request url will be modified in a way that's not observable by page. Must not be
	URL *string `json:"url,omitempty"`
	// If set this allows the request method to be overridden. Must not be set in response to an
	Method *string `json:"method,omitempty"`
	// If set this allows postData to be set. Must not be set in response to an authChallenge.
	PostData *string `json:"postData,omitempty"`
	// If set this allows the request headers to be changed. Must not be set in response to an
	Headers *NetworkHeaders `json:"headers,omitempty"`
	// Response to a requestIntercepted with an authChallenge. Must not be set otherwise.
	AuthChallengeResponse *NetworkAuthChallengeResponse `json:"authChallengeResponse,omitempty"`
}

type NetworkContinueInterceptedRequestResult struct {
}

type NetworkDeleteCookiesParams struct {
	SessionID string `json:"-"`
	// Name of the cookies to remove.
	Name string `json:"name"`
	// If specified, deletes all the cookies with the given name where domain and path match
	URL *string `json:"url,omitempty"`
	// If specified, deletes only cookies with the exact domain.
	Domain *string `json:"domain,omitempty"`
	// If specified, deletes only cookies with the exact path.
	Path *string `json:"path,omitempty"`
	// If specified, deletes only cookies with the the given name and partitionKey where
	PartitionKey *NetworkCookiePartitionKey `json:"partitionKey,omitempty"`
}

type NetworkDeleteCookiesResult struct {
}

type NetworkDisableParams struct {
	SessionID string `json:"-"`
}

type NetworkDisableResult struct {
}

type NetworkEmulateNetworkConditionsParams struct {
	SessionID string `json:"-"`
	// True to emulate internet disconnection.
	Offline bool `json:"offline"`
	// Minimum latency from request sent to response headers received (ms).
	Latency float64 `json:"latency"`
	// Maximal aggregated download throughput (bytes/sec). -1 disables download throttling.
	DownloadThroughput float64 `json:"downloadThroughput"`
	// Maximal aggregated upload throughput (bytes/sec).  -1 disables upload throttling.
	UploadThroughput float64 `json:"uploadThroughput"`
	// Connection type if known.
	ConnectionType *NetworkConnectionType `json:"connectionType,omitempty"`
	// WebRTC packet loss (percent, 0-100). 0 disables packet loss emulation, 100 drops all the packets.
	PacketLoss *float64 `json:"packetLoss,omitempty"`
	// WebRTC packet queue length (packet). 0 removes any queue length limitations.
	PacketQueueLength *int `json:"packetQueueLength,omitempty"`
	// WebRTC packetReordering feature.
	PacketReordering *bool `json:"packetReordering,omitempty"`
}

type NetworkEmulateNetworkConditionsResult struct {
}

type NetworkEmulateNetworkConditionsByRuleParams struct {
	SessionID string `json:"-"`
	// True to emulate internet disconnection.
	Offline bool `json:"offline"`
	// Configure conditions for matching requests. If multiple entries match a request, the first entry wins.  Global
	MatchedNetworkConditions []NetworkNetworkConditions `json:"matchedNetworkConditions"`
}

type NetworkEmulateNetworkConditionsByRuleResult struct {
	// An id for each entry in matchedNetworkConditions. The id will be included in the requestWillBeSentExtraInfo for
	RuleIds []string `json:"ruleIds"`
}

type NetworkOverrideNetworkStateParams struct {
	SessionID string `json:"-"`
	// True to emulate internet disconnection.
	Offline bool `json:"offline"`
	// Minimum latency from request sent to response headers received (ms).
	Latency float64 `json:"latency"`
	// Maximal aggregated download throughput (bytes/sec). -1 disables download throttling.
	DownloadThroughput float64 `json:"downloadThroughput"`
	// Maximal aggregated upload throughput (bytes/sec).  -1 disables upload throttling.
	UploadThroughput float64 `json:"uploadThroughput"`
	// Connection type if known.
	ConnectionType *NetworkConnectionType `json:"connectionType,omitempty"`
}

type NetworkOverrideNetworkStateResult struct {
}

type NetworkEnableParams struct {
	SessionID string `json:"-"`
	// Buffer size in bytes to use when preserving network payloads (XHRs, etc).
	MaxTotalBufferSize *int `json:"maxTotalBufferSize,omitempty"`
	// Per-resource buffer size in bytes to use when preserving network payloads (XHRs, etc).
	MaxResourceBufferSize *int `json:"maxResourceBufferSize,omitempty"`
	// Longest post body size (in bytes) that would be included in requestWillBeSent notification
	MaxPostDataSize *int `json:"maxPostDataSize,omitempty"`
	// Whether DirectSocket chunk send/receive events should be reported.
	ReportDirectSocketTraffic *bool `json:"reportDirectSocketTraffic,omitempty"`
	// Enable storing response bodies outside of renderer, so that these survive
	EnableDurableMessages *bool `json:"enableDurableMessages,omitempty"`
}

type NetworkEnableResult struct {
}

type NetworkConfigureDurableMessagesParams struct {
	SessionID string `json:"-"`
	// Buffer size in bytes to use when preserving network payloads (XHRs, etc).
	MaxTotalBufferSize *int `json:"maxTotalBufferSize,omitempty"`
	// Per-resource buffer size in bytes to use when preserving network payloads (XHRs, etc).
	MaxResourceBufferSize *int `json:"maxResourceBufferSize,omitempty"`
}

type NetworkConfigureDurableMessagesResult struct {
}

type NetworkGetAllCookiesParams struct {
	SessionID string `json:"-"`
}

type NetworkGetAllCookiesResult struct {
	// Array of cookie objects.
	Cookies []NetworkCookie `json:"cookies"`
}

type NetworkGetCertificateParams struct {
	SessionID string `json:"-"`
	// Origin to get certificate for.
	Origin string `json:"origin"`
}

type NetworkGetCertificateResult struct {
	TableNames []string `json:"tableNames"`
}

type NetworkGetCookiesParams struct {
	SessionID string `json:"-"`
	// The list of URLs for which applicable cookies will be fetched.
	Urls []string `json:"urls,omitempty"`
}

type NetworkGetCookiesResult struct {
	// Array of cookie objects.
	Cookies []NetworkCookie `json:"cookies"`
}

type NetworkGetResponseBodyParams struct {
	SessionID string `json:"-"`
	// Identifier of the network request to get content for.
	RequestID NetworkRequestID `json:"requestId"`
}

type NetworkGetResponseBodyResult struct {
	// Response body.
	Body string `json:"body"`
	// True, if content was sent as base64.
	Base64Encoded bool `json:"base64Encoded"`
}

type NetworkGetRequestPostDataParams struct {
	SessionID string `json:"-"`
	// Identifier of the network request to get content for.
	RequestID NetworkRequestID `json:"requestId"`
}

type NetworkGetRequestPostDataResult struct {
	// Request body string, omitting files from multipart requests
	PostData string `json:"postData"`
	// True, if content was sent as base64.
	Base64Encoded bool `json:"base64Encoded"`
}

type NetworkGetResponseBodyForInterceptionParams struct {
	SessionID string `json:"-"`
	// Identifier for the intercepted request to get body for.
	InterceptionID NetworkInterceptionID `json:"interceptionId"`
}

type NetworkGetResponseBodyForInterceptionResult struct {
	// Response body.
	Body string `json:"body"`
	// True, if content was sent as base64.
	Base64Encoded bool `json:"base64Encoded"`
}

type NetworkTakeResponseBodyForInterceptionAsStreamParams struct {
	SessionID      string                `json:"-"`
	InterceptionID NetworkInterceptionID `json:"interceptionId"`
}

type NetworkTakeResponseBodyForInterceptionAsStreamResult struct {
	Stream IOStreamHandle `json:"stream"`
}

type NetworkReplayXHRParams struct {
	SessionID string `json:"-"`
	// Identifier of XHR to replay.
	RequestID NetworkRequestID `json:"requestId"`
}

type NetworkReplayXHRResult struct {
}

type NetworkSearchInResponseBodyParams struct {
	SessionID string `json:"-"`
	// Identifier of the network response to search.
	RequestID NetworkRequestID `json:"requestId"`
	// String to search for.
	Query string `json:"query"`
	// If true, search is case sensitive.
	CaseSensitive *bool `json:"caseSensitive,omitempty"`
	// If true, treats string parameter as regex.
	IsRegex *bool `json:"isRegex,omitempty"`
}

type NetworkSearchInResponseBodyResult struct {
	// List of search matches.
	Result []DebuggerSearchMatch `json:"result"`
}

type NetworkSetBlockedURLsParams struct {
	SessionID string `json:"-"`
	// Patterns to match in the order in which they are given. These patterns
	URLPatterns []NetworkBlockPattern `json:"urlPatterns,omitempty"`
	// URL patterns to block. Wildcards ('*') are allowed.
	Urls []string `json:"urls,omitempty"`
}

type NetworkSetBlockedURLsResult struct {
}

type NetworkSetBypassServiceWorkerParams struct {
	SessionID string `json:"-"`
	// Bypass service worker and load from network.
	Bypass bool `json:"bypass"`
}

type NetworkSetBypassServiceWorkerResult struct {
}

type NetworkSetCacheDisabledParams struct {
	SessionID string `json:"-"`
	// Cache disabled state.
	CacheDisabled bool `json:"cacheDisabled"`
}

type NetworkSetCacheDisabledResult struct {
}

type NetworkSetCookieParams struct {
	SessionID string `json:"-"`
	// Cookie name.
	Name string `json:"name"`
	// Cookie value.
	Value string `json:"value"`
	// The request-URI to associate with the setting of the cookie. This value can affect the
	URL *string `json:"url,omitempty"`
	// Cookie domain.
	Domain *string `json:"domain,omitempty"`
	// Cookie path.
	Path *string `json:"path,omitempty"`
	// True if cookie is secure.
	Secure *bool `json:"secure,omitempty"`
	// True if cookie is http-only.
	HTTPOnly *bool `json:"httpOnly,omitempty"`
	// Cookie SameSite type.
	SameSite *NetworkCookieSameSite `json:"sameSite,omitempty"`
	// Cookie expiration date, session cookie if not set
	Expires *NetworkTimeSinceEpoch `json:"expires,omitempty"`
	// Cookie Priority type.
	Priority *NetworkCookiePriority `json:"priority,omitempty"`
	// Cookie source scheme type.
	SourceScheme *NetworkCookieSourceScheme `json:"sourceScheme,omitempty"`
	// Cookie source port. Valid values are {-1, [1, 65535]}, -1 indicates an unspecified port.
	SourcePort *int `json:"sourcePort,omitempty"`
	// Cookie partition key. If not set, the cookie will be set as not partitioned.
	PartitionKey *NetworkCookiePartitionKey `json:"partitionKey,omitempty"`
}

type NetworkSetCookieResult struct {
	// Always set to true. If an error occurs, the response indicates protocol error.
	Success bool `json:"success"`
}

type NetworkSetCookiesParams struct {
	SessionID string `json:"-"`
	// Cookies to be set.
	Cookies []NetworkCookieParam `json:"cookies"`
}

type NetworkSetCookiesResult struct {
}

type NetworkSetExtraHTTPHeadersParams struct {
	SessionID string `json:"-"`
	// Map with extra HTTP headers.
	Headers NetworkHeaders `json:"headers"`
}

type NetworkSetExtraHTTPHeadersResult struct {
}

type NetworkSetAttachDebugStackParams struct {
	SessionID string `json:"-"`
	// Whether to attach a page script stack for debugging purpose.
	Enabled bool `json:"enabled"`
}

type NetworkSetAttachDebugStackResult struct {
}

type NetworkSetRequestInterceptionParams struct {
	SessionID string `json:"-"`
	// Requests matching any of these patterns will be forwarded and wait for the corresponding
	Patterns []NetworkRequestPattern `json:"patterns"`
}

type NetworkSetRequestInterceptionResult struct {
}

type NetworkSetUserAgentOverrideParams struct {
	SessionID string `json:"-"`
	// User agent to use.
	UserAgent string `json:"userAgent"`
	// Browser language to emulate.
	AcceptLanguage *string `json:"acceptLanguage,omitempty"`
	// The platform navigator.platform should return.
	Platform *string `json:"platform,omitempty"`
	// To be sent in Sec-CH-UA-* headers and returned in navigator.userAgentData
	UserAgentMetadata *EmulationUserAgentMetadata `json:"userAgentMetadata,omitempty"`
}

type NetworkSetUserAgentOverrideResult struct {
}

type NetworkStreamResourceContentParams struct {
	SessionID string `json:"-"`
	// Identifier of the request to stream.
	RequestID NetworkRequestID `json:"requestId"`
}

type NetworkStreamResourceContentResult struct {
	// Data that has been buffered until streaming is enabled. (Encoded as a base64 string when passed over JSON)
	BufferedData string `json:"bufferedData"`
}

type NetworkGetSecurityIsolationStatusParams struct {
	SessionID string `json:"-"`
	// If no frameId is provided, the status of the target is provided.
	FrameID *PageFrameID `json:"frameId,omitempty"`
}

type NetworkGetSecurityIsolationStatusResult struct {
	Status NetworkSecurityIsolationStatus `json:"status"`
}

type NetworkEnableReportingAPIParams struct {
	SessionID string `json:"-"`
	// Whether to enable or disable events for the Reporting API
	Enable bool `json:"enable"`
}

type NetworkEnableReportingAPIResult struct {
}

type NetworkEnableDeviceBoundSessionsParams struct {
	SessionID string `json:"-"`
	// Whether to enable or disable events.
	Enable bool `json:"enable"`
}

type NetworkEnableDeviceBoundSessionsResult struct {
}

type NetworkFetchSchemefulSiteParams struct {
	SessionID string `json:"-"`
	// The URL origin.
	Origin string `json:"origin"`
}

type NetworkFetchSchemefulSiteResult struct {
	// The corresponding schemeful site.
	SchemefulSite string `json:"schemefulSite"`
}

type NetworkLoadNetworkResourceParams struct {
	SessionID string `json:"-"`
	// Frame id to get the resource for. Mandatory for frame targets, and
	FrameID *PageFrameID `json:"frameId,omitempty"`
	// URL of the resource to get content for.
	URL string `json:"url"`
	// Options for the request.
	Options NetworkLoadNetworkResourceOptions `json:"options"`
}

type NetworkLoadNetworkResourceResult struct {
	Resource NetworkLoadNetworkResourcePageResult `json:"resource"`
}

type NetworkSetCookieControlsParams struct {
	SessionID string `json:"-"`
	// Whether 3pc restriction is enabled.
	EnableThirdPartyCookieRestriction bool `json:"enableThirdPartyCookieRestriction"`
	// Whether 3pc grace period exception should be enabled; false by default.
	DisableThirdPartyCookieMetadata bool `json:"disableThirdPartyCookieMetadata"`
	// Whether 3pc heuristics exceptions should be enabled; false by default.
	DisableThirdPartyCookieHeuristics bool `json:"disableThirdPartyCookieHeuristics"`
}

type NetworkSetCookieControlsResult struct {
}

type NetworkDataReceivedEvent struct {
	// Request identifier.
	RequestID NetworkRequestID `json:"requestId"`
	// Timestamp.
	Timestamp NetworkMonotonicTime `json:"timestamp"`
	// Data chunk length.
	DataLength int `json:"dataLength"`
	// Actual bytes received (might be less than dataLength for compressed encodings).
	EncodedDataLength int `json:"encodedDataLength"`
	// Data that was received. (Encoded as a base64 string when passed over JSON)
	Data *string `json:"data,omitempty"`
}

type NetworkEventSourceMessageReceivedEvent struct {
	// Request identifier.
	RequestID NetworkRequestID `json:"requestId"`
	// Timestamp.
	Timestamp NetworkMonotonicTime `json:"timestamp"`
	// Message type.
	EventName string `json:"eventName"`
	// Message identifier.
	EventID string `json:"eventId"`
	// Message content.
	Data string `json:"data"`
}

type NetworkLoadingFailedEvent struct {
	// Request identifier.
	RequestID NetworkRequestID `json:"requestId"`
	// Timestamp.
	Timestamp NetworkMonotonicTime `json:"timestamp"`
	// Resource type.
	Type NetworkResourceType `json:"type"`
	// Error message. List of network errors: https://cs.chromium.org/chromium/src/net/base/net_error_list.h
	ErrorText string `json:"errorText"`
	// True if loading was canceled.
	Canceled *bool `json:"canceled,omitempty"`
	// The reason why loading was blocked, if any.
	BlockedReason *NetworkBlockedReason `json:"blockedReason,omitempty"`
	// The reason why loading was blocked by CORS, if any.
	CorsErrorStatus *NetworkCorsErrorStatus `json:"corsErrorStatus,omitempty"`
}

type NetworkLoadingFinishedEvent struct {
	// Request identifier.
	RequestID NetworkRequestID `json:"requestId"`
	// Timestamp.
	Timestamp NetworkMonotonicTime `json:"timestamp"`
	// Total number of bytes received for this request.
	EncodedDataLength float64 `json:"encodedDataLength"`
}

type NetworkRequestInterceptedEvent struct {
	// Each request the page makes will have a unique id, however if any redirects are encountered
	InterceptionID NetworkInterceptionID `json:"interceptionId"`
	Request        NetworkRequest        `json:"request"`
	// The id of the frame that initiated the request.
	FrameID PageFrameID `json:"frameId"`
	// How the requested resource will be used.
	ResourceType NetworkResourceType `json:"resourceType"`
	// Whether this is a navigation request, which can abort the navigation completely.
	IsNavigationRequest bool `json:"isNavigationRequest"`
	// Set if the request is a navigation that will result in a download.
	IsDownload *bool `json:"isDownload,omitempty"`
	// Redirect location, only sent if a redirect was intercepted.
	RedirectURL *string `json:"redirectUrl,omitempty"`
	// Details of the Authorization Challenge encountered. If this is set then
	AuthChallenge *NetworkAuthChallenge `json:"authChallenge,omitempty"`
	// Response error if intercepted at response stage or if redirect occurred while intercepting
	ResponseErrorReason *NetworkErrorReason `json:"responseErrorReason,omitempty"`
	// Response code if intercepted at response stage or if redirect occurred while intercepting
	ResponseStatusCode *int `json:"responseStatusCode,omitempty"`
	// Response headers if intercepted at the response stage or if redirect occurred while
	ResponseHeaders *NetworkHeaders `json:"responseHeaders,omitempty"`
	// If the intercepted request had a corresponding requestWillBeSent event fired for it, then
	RequestID *NetworkRequestID `json:"requestId,omitempty"`
}

type NetworkRequestServedFromCacheEvent struct {
	// Request identifier.
	RequestID NetworkRequestID `json:"requestId"`
}

type NetworkRequestWillBeSentEvent struct {
	// Request identifier.
	RequestID NetworkRequestID `json:"requestId"`
	// Loader identifier. Empty string if the request is fetched from worker.
	LoaderID NetworkLoaderID `json:"loaderId"`
	// URL of the document this request is loaded for.
	DocumentURL string `json:"documentURL"`
	// Request data.
	Request NetworkRequest `json:"request"`
	// Timestamp.
	Timestamp NetworkMonotonicTime `json:"timestamp"`
	// Timestamp.
	WallTime NetworkTimeSinceEpoch `json:"wallTime"`
	// Request initiator.
	Initiator NetworkInitiator `json:"initiator"`
	// In the case that redirectResponse is populated, this flag indicates whether
	RedirectHasExtraInfo bool `json:"redirectHasExtraInfo"`
	// Redirect response data.
	RedirectResponse *NetworkResponse `json:"redirectResponse,omitempty"`
	// Type of this resource.
	Type *NetworkResourceType `json:"type,omitempty"`
	// Frame identifier.
	FrameID *PageFrameID `json:"frameId,omitempty"`
	// Whether the request is initiated by a user gesture. Defaults to false.
	HasUserGesture *bool `json:"hasUserGesture,omitempty"`
	// The render-blocking behavior of the request.
	RenderBlockingBehavior *NetworkRenderBlockingBehavior `json:"renderBlockingBehavior,omitempty"`
}

type NetworkResourceChangedPriorityEvent struct {
	// Request identifier.
	RequestID NetworkRequestID `json:"requestId"`
	// New priority
	NewPriority NetworkResourcePriority `json:"newPriority"`
	// Timestamp.
	Timestamp NetworkMonotonicTime `json:"timestamp"`
}

type NetworkSignedExchangeReceivedEvent struct {
	// Request identifier.
	RequestID NetworkRequestID `json:"requestId"`
	// Information about the signed exchange response.
	Info NetworkSignedExchangeInfo `json:"info"`
}

type NetworkResponseReceivedEvent struct {
	// Request identifier.
	RequestID NetworkRequestID `json:"requestId"`
	// Loader identifier. Empty string if the request is fetched from worker.
	LoaderID NetworkLoaderID `json:"loaderId"`
	// Timestamp.
	Timestamp NetworkMonotonicTime `json:"timestamp"`
	// Resource type.
	Type NetworkResourceType `json:"type"`
	// Response data.
	Response NetworkResponse `json:"response"`
	// Indicates whether requestWillBeSentExtraInfo and responseReceivedExtraInfo events will be
	HasExtraInfo bool `json:"hasExtraInfo"`
	// Frame identifier.
	FrameID *PageFrameID `json:"frameId,omitempty"`
}

type NetworkWebSocketClosedEvent struct {
	// Request identifier.
	RequestID NetworkRequestID `json:"requestId"`
	// Timestamp.
	Timestamp NetworkMonotonicTime `json:"timestamp"`
}

type NetworkWebSocketCreatedEvent struct {
	// Request identifier.
	RequestID NetworkRequestID `json:"requestId"`
	// WebSocket request URL.
	URL string `json:"url"`
	// Request initiator.
	Initiator *NetworkInitiator `json:"initiator,omitempty"`
}

type NetworkWebSocketFrameErrorEvent struct {
	// Request identifier.
	RequestID NetworkRequestID `json:"requestId"`
	// Timestamp.
	Timestamp NetworkMonotonicTime `json:"timestamp"`
	// WebSocket error message.
	ErrorMessage string `json:"errorMessage"`
}

type NetworkWebSocketFrameReceivedEvent struct {
	// Request identifier.
	RequestID NetworkRequestID `json:"requestId"`
	// Timestamp.
	Timestamp NetworkMonotonicTime `json:"timestamp"`
	// WebSocket response data.
	Response NetworkWebSocketFrame `json:"response"`
}

type NetworkWebSocketFrameSentEvent struct {
	// Request identifier.
	RequestID NetworkRequestID `json:"requestId"`
	// Timestamp.
	Timestamp NetworkMonotonicTime `json:"timestamp"`
	// WebSocket response data.
	Response NetworkWebSocketFrame `json:"response"`
}

type NetworkWebSocketHandshakeResponseReceivedEvent struct {
	// Request identifier.
	RequestID NetworkRequestID `json:"requestId"`
	// Timestamp.
	Timestamp NetworkMonotonicTime `json:"timestamp"`
	// WebSocket response data.
	Response NetworkWebSocketResponse `json:"response"`
}

type NetworkWebSocketWillSendHandshakeRequestEvent struct {
	// Request identifier.
	RequestID NetworkRequestID `json:"requestId"`
	// Timestamp.
	Timestamp NetworkMonotonicTime `json:"timestamp"`
	// UTC Timestamp.
	WallTime NetworkTimeSinceEpoch `json:"wallTime"`
	// WebSocket request data.
	Request NetworkWebSocketRequest `json:"request"`
}

type NetworkWebTransportCreatedEvent struct {
	// WebTransport identifier.
	TransportID NetworkRequestID `json:"transportId"`
	// WebTransport request URL.
	URL string `json:"url"`
	// Timestamp.
	Timestamp NetworkMonotonicTime `json:"timestamp"`
	// Request initiator.
	Initiator *NetworkInitiator `json:"initiator,omitempty"`
}

type NetworkWebTransportConnectionEstablishedEvent struct {
	// WebTransport identifier.
	TransportID NetworkRequestID `json:"transportId"`
	// Timestamp.
	Timestamp NetworkMonotonicTime `json:"timestamp"`
}

type NetworkWebTransportClosedEvent struct {
	// WebTransport identifier.
	TransportID NetworkRequestID `json:"transportId"`
	// Timestamp.
	Timestamp NetworkMonotonicTime `json:"timestamp"`
}

type NetworkDirectTCPSocketCreatedEvent struct {
	Identifier NetworkRequestID `json:"identifier"`
	RemoteAddr string           `json:"remoteAddr"`
	// Unsigned int 16.
	RemotePort int                           `json:"remotePort"`
	Options    NetworkDirectTCPSocketOptions `json:"options"`
	Timestamp  NetworkMonotonicTime          `json:"timestamp"`
	Initiator  *NetworkInitiator             `json:"initiator,omitempty"`
}

type NetworkDirectTCPSocketOpenedEvent struct {
	Identifier NetworkRequestID `json:"identifier"`
	RemoteAddr string           `json:"remoteAddr"`
	// Expected to be unsigned integer.
	RemotePort int                  `json:"remotePort"`
	Timestamp  NetworkMonotonicTime `json:"timestamp"`
	LocalAddr  *string              `json:"localAddr,omitempty"`
	// Expected to be unsigned integer.
	LocalPort *int `json:"localPort,omitempty"`
}

type NetworkDirectTCPSocketAbortedEvent struct {
	Identifier   NetworkRequestID     `json:"identifier"`
	ErrorMessage string               `json:"errorMessage"`
	Timestamp    NetworkMonotonicTime `json:"timestamp"`
}

type NetworkDirectTCPSocketClosedEvent struct {
	Identifier NetworkRequestID     `json:"identifier"`
	Timestamp  NetworkMonotonicTime `json:"timestamp"`
}

type NetworkDirectTCPSocketChunkSentEvent struct {
	Identifier NetworkRequestID     `json:"identifier"`
	Data       string               `json:"data"`
	Timestamp  NetworkMonotonicTime `json:"timestamp"`
}

type NetworkDirectTCPSocketChunkReceivedEvent struct {
	Identifier NetworkRequestID     `json:"identifier"`
	Data       string               `json:"data"`
	Timestamp  NetworkMonotonicTime `json:"timestamp"`
}

type NetworkDirectUDPSocketJoinedMulticastGroupEvent struct {
	Identifier NetworkRequestID `json:"identifier"`
	IPAddress  string           `json:"IPAddress"`
}

type NetworkDirectUDPSocketLeftMulticastGroupEvent struct {
	Identifier NetworkRequestID `json:"identifier"`
	IPAddress  string           `json:"IPAddress"`
}

type NetworkDirectUDPSocketCreatedEvent struct {
	Identifier NetworkRequestID              `json:"identifier"`
	Options    NetworkDirectUDPSocketOptions `json:"options"`
	Timestamp  NetworkMonotonicTime          `json:"timestamp"`
	Initiator  *NetworkInitiator             `json:"initiator,omitempty"`
}

type NetworkDirectUDPSocketOpenedEvent struct {
	Identifier NetworkRequestID `json:"identifier"`
	LocalAddr  string           `json:"localAddr"`
	// Expected to be unsigned integer.
	LocalPort  int                  `json:"localPort"`
	Timestamp  NetworkMonotonicTime `json:"timestamp"`
	RemoteAddr *string              `json:"remoteAddr,omitempty"`
	// Expected to be unsigned integer.
	RemotePort *int `json:"remotePort,omitempty"`
}

type NetworkDirectUDPSocketAbortedEvent struct {
	Identifier   NetworkRequestID     `json:"identifier"`
	ErrorMessage string               `json:"errorMessage"`
	Timestamp    NetworkMonotonicTime `json:"timestamp"`
}

type NetworkDirectUDPSocketClosedEvent struct {
	Identifier NetworkRequestID     `json:"identifier"`
	Timestamp  NetworkMonotonicTime `json:"timestamp"`
}

type NetworkDirectUDPSocketChunkSentEvent struct {
	Identifier NetworkRequestID        `json:"identifier"`
	Message    NetworkDirectUDPMessage `json:"message"`
	Timestamp  NetworkMonotonicTime    `json:"timestamp"`
}

type NetworkDirectUDPSocketChunkReceivedEvent struct {
	Identifier NetworkRequestID        `json:"identifier"`
	Message    NetworkDirectUDPMessage `json:"message"`
	Timestamp  NetworkMonotonicTime    `json:"timestamp"`
}

type NetworkRequestWillBeSentExtraInfoEvent struct {
	// Request identifier. Used to match this information to an existing requestWillBeSent event.
	RequestID NetworkRequestID `json:"requestId"`
	// A list of cookies potentially associated to the requested URL. This includes both cookies sent with
	AssociatedCookies []NetworkAssociatedCookie `json:"associatedCookies"`
	// Raw request headers as they will be sent over the wire.
	Headers NetworkHeaders `json:"headers"`
	// Connection timing information for the request.
	ConnectTiming NetworkConnectTiming `json:"connectTiming"`
	// How the request site's device bound sessions were used during this request.
	DeviceBoundSessionUsages []NetworkDeviceBoundSessionWithUsage `json:"deviceBoundSessionUsages,omitempty"`
	// The client security state set for the request.
	ClientSecurityState *NetworkClientSecurityState `json:"clientSecurityState,omitempty"`
	// Whether the site has partitioned cookies stored in a partition different than the current one.
	SiteHasCookieInOtherPartition *bool `json:"siteHasCookieInOtherPartition,omitempty"`
	// The network conditions id if this request was affected by network conditions configured via
	AppliedNetworkConditionsID *string `json:"appliedNetworkConditionsId,omitempty"`
}

type NetworkResponseReceivedExtraInfoEvent struct {
	// Request identifier. Used to match this information to another responseReceived event.
	RequestID NetworkRequestID `json:"requestId"`
	// A list of cookies which were not stored from the response along with the corresponding
	BlockedCookies []NetworkBlockedSetCookieWithReason `json:"blockedCookies"`
	// Raw response headers as they were received over the wire.
	Headers NetworkHeaders `json:"headers"`
	// The IP address space of the resource. The address space can only be determined once the transport
	ResourceIPAddressSpace NetworkIPAddressSpace `json:"resourceIPAddressSpace"`
	// The status code of the response. This is useful in cases the request failed and no responseReceived
	StatusCode int `json:"statusCode"`
	// Raw response header text as it was received over the wire. The raw text may not always be
	HeadersText *string `json:"headersText,omitempty"`
	// The cookie partition key that will be used to store partitioned cookies set in this response.
	CookiePartitionKey *NetworkCookiePartitionKey `json:"cookiePartitionKey,omitempty"`
	// True if partitioned cookies are enabled, but the partition key is not serializable to string.
	CookiePartitionKeyOpaque *bool `json:"cookiePartitionKeyOpaque,omitempty"`
	// A list of cookies which should have been blocked by 3PCD but are exempted and stored from
	ExemptedCookies []NetworkExemptedSetCookieWithReason `json:"exemptedCookies,omitempty"`
}

type NetworkResponseReceivedEarlyHintsEvent struct {
	// Request identifier. Used to match this information to another responseReceived event.
	RequestID NetworkRequestID `json:"requestId"`
	// Raw response headers as they were received over the wire.
	Headers NetworkHeaders `json:"headers"`
}

type NetworkTrustTokenOperationDoneEvent struct {
	// Detailed success or error status of the operation.
	Status    string                         `json:"status"`
	Type      NetworkTrustTokenOperationType `json:"type"`
	RequestID NetworkRequestID               `json:"requestId"`
	// Top level origin. The context in which the operation was attempted.
	TopLevelOrigin *string `json:"topLevelOrigin,omitempty"`
	// Origin of the issuer in case of a "Issuance" or "Redemption" operation.
	IssuerOrigin *string `json:"issuerOrigin,omitempty"`
	// The number of obtained Trust Tokens on a successful "Issuance" operation.
	IssuedTokenCount *int `json:"issuedTokenCount,omitempty"`
}

type NetworkPolicyUpdatedEvent struct {
}

type NetworkReportingAPIReportAddedEvent struct {
	Report NetworkReportingAPIReport `json:"report"`
}

type NetworkReportingAPIReportUpdatedEvent struct {
	Report NetworkReportingAPIReport `json:"report"`
}

type NetworkReportingAPIEndpointsChangedForOriginEvent struct {
	// Origin of the document(s) which configured the endpoints.
	Origin    string                        `json:"origin"`
	Endpoints []NetworkReportingAPIEndpoint `json:"endpoints"`
}

type NetworkDeviceBoundSessionsAddedEvent struct {
	// The device bound sessions.
	Sessions []NetworkDeviceBoundSession `json:"sessions"`
}

type NetworkDeviceBoundSessionEventOccurredEvent struct {
	// A unique identifier for this session event.
	EventID NetworkDeviceBoundSessionEventID `json:"eventId"`
	// The site this session event is associated with.
	Site string `json:"site"`
	// Whether this event was considered successful.
	Succeeded bool `json:"succeeded"`
	// The session ID this event is associated with. May not be populated for
	SessionID *string `json:"sessionId,omitempty"`
	// The below are the different session event type details. Exactly one is populated.
	CreationEventDetails    *NetworkCreationEventDetails    `json:"creationEventDetails,omitempty"`
	RefreshEventDetails     *NetworkRefreshEventDetails     `json:"refreshEventDetails,omitempty"`
	TerminationEventDetails *NetworkTerminationEventDetails `json:"terminationEventDetails,omitempty"`
	ChallengeEventDetails   *NetworkChallengeEventDetails   `json:"challengeEventDetails,omitempty"`
}

type OverlaySourceOrderConfig struct {
	// the color to outline the given element in.
	ParentOutlineColor DOMRGBA `json:"parentOutlineColor"`
	// the color to outline the child elements in.
	ChildOutlineColor DOMRGBA `json:"childOutlineColor"`
}

type OverlayGridHighlightConfig struct {
	// Whether the extension lines from grid cells to the rulers should be shown (default: false).
	ShowGridExtensionLines *bool `json:"showGridExtensionLines,omitempty"`
	// Show Positive line number labels (default: false).
	ShowPositiveLineNumbers *bool `json:"showPositiveLineNumbers,omitempty"`
	// Show Negative line number labels (default: false).
	ShowNegativeLineNumbers *bool `json:"showNegativeLineNumbers,omitempty"`
	// Show area name labels (default: false).
	ShowAreaNames *bool `json:"showAreaNames,omitempty"`
	// Show line name labels (default: false).
	ShowLineNames *bool `json:"showLineNames,omitempty"`
	// Show track size labels (default: false).
	ShowTrackSizes *bool `json:"showTrackSizes,omitempty"`
	// The grid container border highlight color (default: transparent).
	GridBorderColor *DOMRGBA `json:"gridBorderColor,omitempty"`
	// The cell border color (default: transparent). Deprecated, please use rowLineColor and columnLineColor instead.
	CellBorderColor *DOMRGBA `json:"cellBorderColor,omitempty"`
	// The row line color (default: transparent).
	RowLineColor *DOMRGBA `json:"rowLineColor,omitempty"`
	// The column line color (default: transparent).
	ColumnLineColor *DOMRGBA `json:"columnLineColor,omitempty"`
	// Whether the grid border is dashed (default: false).
	GridBorderDash *bool `json:"gridBorderDash,omitempty"`
	// Whether the cell border is dashed (default: false). Deprecated, please us rowLineDash and columnLineDash instead.
	CellBorderDash *bool `json:"cellBorderDash,omitempty"`
	// Whether row lines are dashed (default: false).
	RowLineDash *bool `json:"rowLineDash,omitempty"`
	// Whether column lines are dashed (default: false).
	ColumnLineDash *bool `json:"columnLineDash,omitempty"`
	// The row gap highlight fill color (default: transparent).
	RowGapColor *DOMRGBA `json:"rowGapColor,omitempty"`
	// The row gap hatching fill color (default: transparent).
	RowHatchColor *DOMRGBA `json:"rowHatchColor,omitempty"`
	// The column gap highlight fill color (default: transparent).
	ColumnGapColor *DOMRGBA `json:"columnGapColor,omitempty"`
	// The column gap hatching fill color (default: transparent).
	ColumnHatchColor *DOMRGBA `json:"columnHatchColor,omitempty"`
	// The named grid areas border color (Default: transparent).
	AreaBorderColor *DOMRGBA `json:"areaBorderColor,omitempty"`
	// The grid container background color (Default: transparent).
	GridBackgroundColor *DOMRGBA `json:"gridBackgroundColor,omitempty"`
}

type OverlayFlexContainerHighlightConfig struct {
	// The style of the container border
	ContainerBorder *OverlayLineStyle `json:"containerBorder,omitempty"`
	// The style of the separator between lines
	LineSeparator *OverlayLineStyle `json:"lineSeparator,omitempty"`
	// The style of the separator between items
	ItemSeparator *OverlayLineStyle `json:"itemSeparator,omitempty"`
	// Style of content-distribution space on the main axis (justify-content).
	MainDistributedSpace *OverlayBoxStyle `json:"mainDistributedSpace,omitempty"`
	// Style of content-distribution space on the cross axis (align-content).
	CrossDistributedSpace *OverlayBoxStyle `json:"crossDistributedSpace,omitempty"`
	// Style of empty space caused by row gaps (gap/row-gap).
	RowGapSpace *OverlayBoxStyle `json:"rowGapSpace,omitempty"`
	// Style of empty space caused by columns gaps (gap/column-gap).
	ColumnGapSpace *OverlayBoxStyle `json:"columnGapSpace,omitempty"`
	// Style of the self-alignment line (align-items).
	CrossAlignment *OverlayLineStyle `json:"crossAlignment,omitempty"`
}

type OverlayFlexItemHighlightConfig struct {
	// Style of the box representing the item's base size
	BaseSizeBox *OverlayBoxStyle `json:"baseSizeBox,omitempty"`
	// Style of the border around the box representing the item's base size
	BaseSizeBorder *OverlayLineStyle `json:"baseSizeBorder,omitempty"`
	// Style of the arrow representing if the item grew or shrank
	FlexibilityArrow *OverlayLineStyle `json:"flexibilityArrow,omitempty"`
}

type OverlayLineStyle struct {
	// The color of the line (default: transparent)
	Color *DOMRGBA `json:"color,omitempty"`
	// The line pattern (default: solid)
	Pattern *string `json:"pattern,omitempty"`
}

type OverlayBoxStyle struct {
	// The background color for the box (default: transparent)
	FillColor *DOMRGBA `json:"fillColor,omitempty"`
	// The hatching color for the box (default: transparent)
	HatchColor *DOMRGBA `json:"hatchColor,omitempty"`
}

type OverlayContrastAlgorithm string

type OverlayHighlightConfig struct {
	// Whether the node info tooltip should be shown (default: false).
	ShowInfo *bool `json:"showInfo,omitempty"`
	// Whether the node styles in the tooltip (default: false).
	ShowStyles *bool `json:"showStyles,omitempty"`
	// Whether the rulers should be shown (default: false).
	ShowRulers *bool `json:"showRulers,omitempty"`
	// Whether the a11y info should be shown (default: true).
	ShowAccessibilityInfo *bool `json:"showAccessibilityInfo,omitempty"`
	// Whether the extension lines from node to the rulers should be shown (default: false).
	ShowExtensionLines *bool `json:"showExtensionLines,omitempty"`
	// The content box highlight fill color (default: transparent).
	ContentColor *DOMRGBA `json:"contentColor,omitempty"`
	// The padding highlight fill color (default: transparent).
	PaddingColor *DOMRGBA `json:"paddingColor,omitempty"`
	// The border highlight fill color (default: transparent).
	BorderColor *DOMRGBA `json:"borderColor,omitempty"`
	// The margin highlight fill color (default: transparent).
	MarginColor *DOMRGBA `json:"marginColor,omitempty"`
	// The event target element highlight fill color (default: transparent).
	EventTargetColor *DOMRGBA `json:"eventTargetColor,omitempty"`
	// The shape outside fill color (default: transparent).
	ShapeColor *DOMRGBA `json:"shapeColor,omitempty"`
	// The shape margin fill color (default: transparent).
	ShapeMarginColor *DOMRGBA `json:"shapeMarginColor,omitempty"`
	// The grid layout color (default: transparent).
	CSSGridColor *DOMRGBA `json:"cssGridColor,omitempty"`
	// The color format used to format color styles (default: hex).
	ColorFormat *OverlayColorFormat `json:"colorFormat,omitempty"`
	// The grid layout highlight configuration (default: all transparent).
	GridHighlightConfig *OverlayGridHighlightConfig `json:"gridHighlightConfig,omitempty"`
	// The flex container highlight configuration (default: all transparent).
	FlexContainerHighlightConfig *OverlayFlexContainerHighlightConfig `json:"flexContainerHighlightConfig,omitempty"`
	// The flex item highlight configuration (default: all transparent).
	FlexItemHighlightConfig *OverlayFlexItemHighlightConfig `json:"flexItemHighlightConfig,omitempty"`
	// The contrast algorithm to use for the contrast ratio (default: aa).
	ContrastAlgorithm *OverlayContrastAlgorithm `json:"contrastAlgorithm,omitempty"`
	// The container query container highlight configuration (default: all transparent).
	ContainerQueryContainerHighlightConfig *OverlayContainerQueryContainerHighlightConfig `json:"containerQueryContainerHighlightConfig,omitempty"`
}

type OverlayColorFormat string

type OverlayGridNodeHighlightConfig struct {
	// A descriptor for the highlight appearance.
	GridHighlightConfig OverlayGridHighlightConfig `json:"gridHighlightConfig"`
	// Identifier of the node to highlight.
	NodeID DOMNodeID `json:"nodeId"`
}

type OverlayFlexNodeHighlightConfig struct {
	// A descriptor for the highlight appearance of flex containers.
	FlexContainerHighlightConfig OverlayFlexContainerHighlightConfig `json:"flexContainerHighlightConfig"`
	// Identifier of the node to highlight.
	NodeID DOMNodeID `json:"nodeId"`
}

type OverlayScrollSnapContainerHighlightConfig struct {
	// The style of the snapport border (default: transparent)
	SnapportBorder *OverlayLineStyle `json:"snapportBorder,omitempty"`
	// The style of the snap area border (default: transparent)
	SnapAreaBorder *OverlayLineStyle `json:"snapAreaBorder,omitempty"`
	// The margin highlight fill color (default: transparent).
	ScrollMarginColor *DOMRGBA `json:"scrollMarginColor,omitempty"`
	// The padding highlight fill color (default: transparent).
	ScrollPaddingColor *DOMRGBA `json:"scrollPaddingColor,omitempty"`
}

type OverlayScrollSnapHighlightConfig struct {
	// A descriptor for the highlight appearance of scroll snap containers.
	ScrollSnapContainerHighlightConfig OverlayScrollSnapContainerHighlightConfig `json:"scrollSnapContainerHighlightConfig"`
	// Identifier of the node to highlight.
	NodeID DOMNodeID `json:"nodeId"`
}

type OverlayHingeConfig struct {
	// A rectangle represent hinge
	Rect DOMRect `json:"rect"`
	// The content box highlight fill color (default: a dark color).
	ContentColor *DOMRGBA `json:"contentColor,omitempty"`
	// The content box highlight outline color (default: transparent).
	OutlineColor *DOMRGBA `json:"outlineColor,omitempty"`
}

type OverlayWindowControlsOverlayConfig struct {
	// Whether the title bar CSS should be shown when emulating the Window Controls Overlay.
	ShowCSS bool `json:"showCSS"`
	// Selected platforms to show the overlay.
	SelectedPlatform string `json:"selectedPlatform"`
	// The theme color defined in app manifest.
	ThemeColor string `json:"themeColor"`
}

type OverlayContainerQueryHighlightConfig struct {
	// A descriptor for the highlight appearance of container query containers.
	ContainerQueryContainerHighlightConfig OverlayContainerQueryContainerHighlightConfig `json:"containerQueryContainerHighlightConfig"`
	// Identifier of the container node to highlight.
	NodeID DOMNodeID `json:"nodeId"`
}

type OverlayContainerQueryContainerHighlightConfig struct {
	// The style of the container border.
	ContainerBorder *OverlayLineStyle `json:"containerBorder,omitempty"`
	// The style of the descendants' borders.
	DescendantBorder *OverlayLineStyle `json:"descendantBorder,omitempty"`
}

type OverlayIsolatedElementHighlightConfig struct {
	// A descriptor for the highlight appearance of an element in isolation mode.
	IsolationModeHighlightConfig OverlayIsolationModeHighlightConfig `json:"isolationModeHighlightConfig"`
	// Identifier of the isolated element to highlight.
	NodeID DOMNodeID `json:"nodeId"`
}

type OverlayIsolationModeHighlightConfig struct {
	// The fill color of the resizers (default: transparent).
	ResizerColor *DOMRGBA `json:"resizerColor,omitempty"`
	// The fill color for resizer handles (default: transparent).
	ResizerHandleColor *DOMRGBA `json:"resizerHandleColor,omitempty"`
	// The fill color for the mask covering non-isolated elements (default: transparent).
	MaskColor *DOMRGBA `json:"maskColor,omitempty"`
}

type OverlayInspectMode string

type OverlayInspectedElementAnchorConfig struct {
	// Identifier of the node to highlight.
	NodeID *DOMNodeID `json:"nodeId,omitempty"`
	// Identifier of the backend node to highlight.
	BackendNodeID *DOMBackendNodeID `json:"backendNodeId,omitempty"`
}

type OverlayDisableParams struct {
	SessionID string `json:"-"`
}

type OverlayDisableResult struct {
}

type OverlayEnableParams struct {
	SessionID string `json:"-"`
}

type OverlayEnableResult struct {
}

type OverlayGetHighlightObjectForTestParams struct {
	SessionID string `json:"-"`
	// Id of the node to get highlight object for.
	NodeID DOMNodeID `json:"nodeId"`
	// Whether to include distance info.
	IncludeDistance *bool `json:"includeDistance,omitempty"`
	// Whether to include style info.
	IncludeStyle *bool `json:"includeStyle,omitempty"`
	// The color format to get config with (default: hex).
	ColorFormat *OverlayColorFormat `json:"colorFormat,omitempty"`
	// Whether to show accessibility info (default: true).
	ShowAccessibilityInfo *bool `json:"showAccessibilityInfo,omitempty"`
}

type OverlayGetHighlightObjectForTestResult struct {
	// Highlight data for the node.
	Highlight map[string]any `json:"highlight"`
}

type OverlayGetGridHighlightObjectsForTestParams struct {
	SessionID string `json:"-"`
	// Ids of the node to get highlight object for.
	NodeIds []DOMNodeID `json:"nodeIds"`
}

type OverlayGetGridHighlightObjectsForTestResult struct {
	// Grid Highlight data for the node ids provided.
	Highlights map[string]any `json:"highlights"`
}

type OverlayGetSourceOrderHighlightObjectForTestParams struct {
	SessionID string `json:"-"`
	// Id of the node to highlight.
	NodeID DOMNodeID `json:"nodeId"`
}

type OverlayGetSourceOrderHighlightObjectForTestResult struct {
	// Source order highlight data for the node id provided.
	Highlight map[string]any `json:"highlight"`
}

type OverlayHideHighlightParams struct {
	SessionID string `json:"-"`
}

type OverlayHideHighlightResult struct {
}

type OverlayHighlightFrameParams struct {
	SessionID string `json:"-"`
	// Identifier of the frame to highlight.
	FrameID PageFrameID `json:"frameId"`
	// The content box highlight fill color (default: transparent).
	ContentColor *DOMRGBA `json:"contentColor,omitempty"`
	// The content box highlight outline color (default: transparent).
	ContentOutlineColor *DOMRGBA `json:"contentOutlineColor,omitempty"`
}

type OverlayHighlightFrameResult struct {
}

type OverlayHighlightNodeParams struct {
	SessionID string `json:"-"`
	// A descriptor for the highlight appearance.
	HighlightConfig OverlayHighlightConfig `json:"highlightConfig"`
	// Identifier of the node to highlight.
	NodeID *DOMNodeID `json:"nodeId,omitempty"`
	// Identifier of the backend node to highlight.
	BackendNodeID *DOMBackendNodeID `json:"backendNodeId,omitempty"`
	// JavaScript object id of the node to be highlighted.
	ObjectID *RuntimeRemoteObjectID `json:"objectId,omitempty"`
	// Selectors to highlight relevant nodes.
	Selector *string `json:"selector,omitempty"`
}

type OverlayHighlightNodeResult struct {
}

type OverlayHighlightQuadParams struct {
	SessionID string `json:"-"`
	// Quad to highlight
	Quad DOMQuad `json:"quad"`
	// The highlight fill color (default: transparent).
	Color *DOMRGBA `json:"color,omitempty"`
	// The highlight outline color (default: transparent).
	OutlineColor *DOMRGBA `json:"outlineColor,omitempty"`
}

type OverlayHighlightQuadResult struct {
}

type OverlayHighlightRectParams struct {
	SessionID string `json:"-"`
	// X coordinate
	X int `json:"x"`
	// Y coordinate
	Y int `json:"y"`
	// Rectangle width
	Width int `json:"width"`
	// Rectangle height
	Height int `json:"height"`
	// The highlight fill color (default: transparent).
	Color *DOMRGBA `json:"color,omitempty"`
	// The highlight outline color (default: transparent).
	OutlineColor *DOMRGBA `json:"outlineColor,omitempty"`
}

type OverlayHighlightRectResult struct {
}

type OverlayHighlightSourceOrderParams struct {
	SessionID string `json:"-"`
	// A descriptor for the appearance of the overlay drawing.
	SourceOrderConfig OverlaySourceOrderConfig `json:"sourceOrderConfig"`
	// Identifier of the node to highlight.
	NodeID *DOMNodeID `json:"nodeId,omitempty"`
	// Identifier of the backend node to highlight.
	BackendNodeID *DOMBackendNodeID `json:"backendNodeId,omitempty"`
	// JavaScript object id of the node to be highlighted.
	ObjectID *RuntimeRemoteObjectID `json:"objectId,omitempty"`
}

type OverlayHighlightSourceOrderResult struct {
}

type OverlaySetInspectModeParams struct {
	SessionID string `json:"-"`
	// Set an inspection mode.
	Mode OverlayInspectMode `json:"mode"`
	// A descriptor for the highlight appearance of hovered-over nodes. May be omitted if `enabled
	HighlightConfig *OverlayHighlightConfig `json:"highlightConfig,omitempty"`
}

type OverlaySetInspectModeResult struct {
}

type OverlaySetShowAdHighlightsParams struct {
	SessionID string `json:"-"`
	// True for showing ad highlights
	Show bool `json:"show"`
}

type OverlaySetShowAdHighlightsResult struct {
}

type OverlaySetPausedInDebuggerMessageParams struct {
	SessionID string `json:"-"`
	// The message to display, also triggers resume and step over controls.
	Message *string `json:"message,omitempty"`
}

type OverlaySetPausedInDebuggerMessageResult struct {
}

type OverlaySetShowDebugBordersParams struct {
	SessionID string `json:"-"`
	// True for showing debug borders
	Show bool `json:"show"`
}

type OverlaySetShowDebugBordersResult struct {
}

type OverlaySetShowFPSCounterParams struct {
	SessionID string `json:"-"`
	// True for showing the FPS counter
	Show bool `json:"show"`
}

type OverlaySetShowFPSCounterResult struct {
}

type OverlaySetShowGridOverlaysParams struct {
	SessionID string `json:"-"`
	// An array of node identifiers and descriptors for the highlight appearance.
	GridNodeHighlightConfigs []OverlayGridNodeHighlightConfig `json:"gridNodeHighlightConfigs"`
}

type OverlaySetShowGridOverlaysResult struct {
}

type OverlaySetShowFlexOverlaysParams struct {
	SessionID string `json:"-"`
	// An array of node identifiers and descriptors for the highlight appearance.
	FlexNodeHighlightConfigs []OverlayFlexNodeHighlightConfig `json:"flexNodeHighlightConfigs"`
}

type OverlaySetShowFlexOverlaysResult struct {
}

type OverlaySetShowScrollSnapOverlaysParams struct {
	SessionID string `json:"-"`
	// An array of node identifiers and descriptors for the highlight appearance.
	ScrollSnapHighlightConfigs []OverlayScrollSnapHighlightConfig `json:"scrollSnapHighlightConfigs"`
}

type OverlaySetShowScrollSnapOverlaysResult struct {
}

type OverlaySetShowContainerQueryOverlaysParams struct {
	SessionID string `json:"-"`
	// An array of node identifiers and descriptors for the highlight appearance.
	ContainerQueryHighlightConfigs []OverlayContainerQueryHighlightConfig `json:"containerQueryHighlightConfigs"`
}

type OverlaySetShowContainerQueryOverlaysResult struct {
}

type OverlaySetShowInspectedElementAnchorParams struct {
	SessionID string `json:"-"`
	// Node identifier for which to show an anchor for.
	InspectedElementAnchorConfig OverlayInspectedElementAnchorConfig `json:"inspectedElementAnchorConfig"`
}

type OverlaySetShowInspectedElementAnchorResult struct {
}

type OverlaySetShowPaintRectsParams struct {
	SessionID string `json:"-"`
	// True for showing paint rectangles
	Result bool `json:"result"`
}

type OverlaySetShowPaintRectsResult struct {
}

type OverlaySetShowLayoutShiftRegionsParams struct {
	SessionID string `json:"-"`
	// True for showing layout shift regions
	Result bool `json:"result"`
}

type OverlaySetShowLayoutShiftRegionsResult struct {
}

type OverlaySetShowScrollBottleneckRectsParams struct {
	SessionID string `json:"-"`
	// True for showing scroll bottleneck rects
	Show bool `json:"show"`
}

type OverlaySetShowScrollBottleneckRectsResult struct {
}

type OverlaySetShowHitTestBordersParams struct {
	SessionID string `json:"-"`
	// True for showing hit-test borders
	Show bool `json:"show"`
}

type OverlaySetShowHitTestBordersResult struct {
}

type OverlaySetShowWebVitalsParams struct {
	SessionID string `json:"-"`
	Show      bool   `json:"show"`
}

type OverlaySetShowWebVitalsResult struct {
}

type OverlaySetShowViewportSizeOnResizeParams struct {
	SessionID string `json:"-"`
	// Whether to paint size or not.
	Show bool `json:"show"`
}

type OverlaySetShowViewportSizeOnResizeResult struct {
}

type OverlaySetShowHingeParams struct {
	SessionID string `json:"-"`
	// hinge data, null means hideHinge
	HingeConfig *OverlayHingeConfig `json:"hingeConfig,omitempty"`
}

type OverlaySetShowHingeResult struct {
}

type OverlaySetShowIsolatedElementsParams struct {
	SessionID string `json:"-"`
	// An array of node identifiers and descriptors for the highlight appearance.
	IsolatedElementHighlightConfigs []OverlayIsolatedElementHighlightConfig `json:"isolatedElementHighlightConfigs"`
}

type OverlaySetShowIsolatedElementsResult struct {
}

type OverlaySetShowWindowControlsOverlayParams struct {
	SessionID string `json:"-"`
	// Window Controls Overlay data, null means hide Window Controls Overlay
	WindowControlsOverlayConfig *OverlayWindowControlsOverlayConfig `json:"windowControlsOverlayConfig,omitempty"`
}

type OverlaySetShowWindowControlsOverlayResult struct {
}

type OverlayInspectNodeRequestedEvent struct {
	// Id of the node to inspect.
	BackendNodeID DOMBackendNodeID `json:"backendNodeId"`
}

type OverlayNodeHighlightRequestedEvent struct {
	NodeID DOMNodeID `json:"nodeId"`
}

type OverlayScreenshotRequestedEvent struct {
	// Viewport to capture, in device independent pixels (dip).
	Viewport PageViewport `json:"viewport"`
}

type OverlayInspectPanelShowRequestedEvent struct {
	// Id of the node to show in the panel.
	BackendNodeID DOMBackendNodeID `json:"backendNodeId"`
}

type OverlayInspectedElementWindowRestoredEvent struct {
	// Id of the node to restore the floating window for.
	BackendNodeID DOMBackendNodeID `json:"backendNodeId"`
}

type OverlayInspectModeCanceledEvent struct {
}

type PWAFileHandlerAccept struct {
	// New name of the mimetype according to
	MediaType      string   `json:"mediaType"`
	FileExtensions []string `json:"fileExtensions"`
}

type PWAFileHandler struct {
	Action      string                 `json:"action"`
	Accepts     []PWAFileHandlerAccept `json:"accepts"`
	DisplayName string                 `json:"displayName"`
}

type PWADisplayMode string

type PWAGetOsAppStateParams struct {
	SessionID string `json:"-"`
	// The id from the webapp's manifest file, commonly it's the url of the
	ManifestID string `json:"manifestId"`
}

type PWAGetOsAppStateResult struct {
	BadgeCount   int              `json:"badgeCount"`
	FileHandlers []PWAFileHandler `json:"fileHandlers"`
}

type PWAInstallParams struct {
	SessionID  string `json:"-"`
	ManifestID string `json:"manifestId"`
	// The location of the app or bundle overriding the one derived from the
	InstallURLOrBundleURL *string `json:"installUrlOrBundleUrl,omitempty"`
}

type PWAInstallResult struct {
}

type PWAUninstallParams struct {
	SessionID  string `json:"-"`
	ManifestID string `json:"manifestId"`
}

type PWAUninstallResult struct {
}

type PWALaunchParams struct {
	SessionID  string  `json:"-"`
	ManifestID string  `json:"manifestId"`
	URL        *string `json:"url,omitempty"`
}

type PWALaunchResult struct {
	// ID of the tab target created as a result.
	TargetID TargetTargetID `json:"targetId"`
}

type PWALaunchFilesInAppParams struct {
	SessionID  string   `json:"-"`
	ManifestID string   `json:"manifestId"`
	Files      []string `json:"files"`
}

type PWALaunchFilesInAppResult struct {
	// IDs of the tab targets created as the result.
	TargetIds []TargetTargetID `json:"targetIds"`
}

type PWAOpenCurrentPageInAppParams struct {
	SessionID  string `json:"-"`
	ManifestID string `json:"manifestId"`
}

type PWAOpenCurrentPageInAppResult struct {
}

type PWAChangeAppUserSettingsParams struct {
	SessionID  string `json:"-"`
	ManifestID string `json:"manifestId"`
	// If user allows the links clicked on by the user in the app's scope, or
	LinkCapturing *bool           `json:"linkCapturing,omitempty"`
	DisplayMode   *PWADisplayMode `json:"displayMode,omitempty"`
}

type PWAChangeAppUserSettingsResult struct {
}

type PageFrameID string

type PageAdFrameType string

type PageAdFrameExplanation string

type PageAdFrameStatus struct {
	AdFrameType  PageAdFrameType          `json:"adFrameType"`
	Explanations []PageAdFrameExplanation `json:"explanations,omitempty"`
}

type PageAdScriptID struct {
	// Script Id of the script which caused a script or frame to be labelled as
	ScriptID RuntimeScriptID `json:"scriptId"`
	// Id of scriptId's debugger.
	DebuggerID RuntimeUniqueDebuggerID `json:"debuggerId"`
}

type PageAdScriptAncestry struct {
	// A chain of `AdScriptId`s representing the ancestry of an ad script that
	AncestryChain []PageAdScriptID `json:"ancestryChain"`
	// The filterlist rule that caused the root (last) script in
	RootScriptFilterlistRule *string `json:"rootScriptFilterlistRule,omitempty"`
}

type PageSecureContextType string

type PageCrossOriginIsolatedContextType string

type PageGatedAPIFeatures string

type PagePermissionsPolicyFeature string

type PagePermissionsPolicyBlockReason string

type PagePermissionsPolicyBlockLocator struct {
	FrameID     PageFrameID                      `json:"frameId"`
	BlockReason PagePermissionsPolicyBlockReason `json:"blockReason"`
}

type PagePermissionsPolicyFeatureState struct {
	Feature PagePermissionsPolicyFeature       `json:"feature"`
	Allowed bool                               `json:"allowed"`
	Locator *PagePermissionsPolicyBlockLocator `json:"locator,omitempty"`
}

type PageOriginTrialTokenStatus string

type PageOriginTrialStatus string

type PageOriginTrialUsageRestriction string

type PageOriginTrialToken struct {
	Origin           string                          `json:"origin"`
	MatchSubDomains  bool                            `json:"matchSubDomains"`
	TrialName        string                          `json:"trialName"`
	ExpiryTime       NetworkTimeSinceEpoch           `json:"expiryTime"`
	IsThirdParty     bool                            `json:"isThirdParty"`
	UsageRestriction PageOriginTrialUsageRestriction `json:"usageRestriction"`
}

type PageOriginTrialTokenWithStatus struct {
	RawTokenText string `json:"rawTokenText"`
	// `parsedToken` is present only when the token is extractable and
	ParsedToken *PageOriginTrialToken      `json:"parsedToken,omitempty"`
	Status      PageOriginTrialTokenStatus `json:"status"`
}

type PageOriginTrial struct {
	TrialName        string                           `json:"trialName"`
	Status           PageOriginTrialStatus            `json:"status"`
	TokensWithStatus []PageOriginTrialTokenWithStatus `json:"tokensWithStatus"`
}

type PageSecurityOriginDetails struct {
	// Indicates whether the frame document's security origin is one
	IsLocalhost bool `json:"isLocalhost"`
}

type PageFrame struct {
	// Frame unique identifier.
	ID PageFrameID `json:"id"`
	// Parent frame identifier.
	ParentID *PageFrameID `json:"parentId,omitempty"`
	// Identifier of the loader associated with this frame.
	LoaderID NetworkLoaderID `json:"loaderId"`
	// Frame's name as specified in the tag.
	Name *string `json:"name,omitempty"`
	// Frame document's URL without fragment.
	URL string `json:"url"`
	// Frame document's URL fragment including the '#'.
	URLFragment *string `json:"urlFragment,omitempty"`
	// Frame document's registered domain, taking the public suffixes list into account.
	DomainAndRegistry string `json:"domainAndRegistry"`
	// Frame document's security origin.
	SecurityOrigin string `json:"securityOrigin"`
	// Additional details about the frame document's security origin.
	SecurityOriginDetails *PageSecurityOriginDetails `json:"securityOriginDetails,omitempty"`
	// Frame document's mimeType as determined by the browser.
	MimeType string `json:"mimeType"`
	// If the frame failed to load, this contains the URL that could not be loaded. Note that unlike url above, this URL may contain a fragment.
	UnreachableURL *string `json:"unreachableUrl,omitempty"`
	// Indicates whether this frame was tagged as an ad and why.
	AdFrameStatus *PageAdFrameStatus `json:"adFrameStatus,omitempty"`
	// Indicates whether the main document is a secure context and explains why that is the case.
	SecureContextType PageSecureContextType `json:"secureContextType"`
	// Indicates whether this is a cross origin isolated context.
	CrossOriginIsolatedContextType PageCrossOriginIsolatedContextType `json:"crossOriginIsolatedContextType"`
	// Indicated which gated APIs / features are available.
	GatedAPIFeatures []PageGatedAPIFeatures `json:"gatedAPIFeatures"`
}

type PageFrameResource struct {
	// Resource URL.
	URL string `json:"url"`
	// Type of this resource.
	Type NetworkResourceType `json:"type"`
	// Resource mimeType as determined by the browser.
	MimeType string `json:"mimeType"`
	// last-modified timestamp as reported by server.
	LastModified *NetworkTimeSinceEpoch `json:"lastModified,omitempty"`
	// Resource content size.
	ContentSize *float64 `json:"contentSize,omitempty"`
	// True if the resource failed to load.
	Failed *bool `json:"failed,omitempty"`
	// True if the resource was canceled during loading.
	Canceled *bool `json:"canceled,omitempty"`
}

type PageFrameResourceTree struct {
	// Frame information for this tree item.
	Frame PageFrame `json:"frame"`
	// Child frames.
	ChildFrames []PageFrameResourceTree `json:"childFrames,omitempty"`
	// Information about frame resources.
	Resources []PageFrameResource `json:"resources"`
}

type PageFrameTree struct {
	// Frame information for this tree item.
	Frame PageFrame `json:"frame"`
	// Child frames.
	ChildFrames []PageFrameTree `json:"childFrames,omitempty"`
}

type PageScriptIdentifier string

type PageTransitionType string

type PageNavigationEntry struct {
	// Unique id of the navigation history entry.
	ID int `json:"id"`
	// URL of the navigation history entry.
	URL string `json:"url"`
	// URL that the user typed in the url bar.
	UserTypedURL string `json:"userTypedURL"`
	// Title of the navigation history entry.
	Title string `json:"title"`
	// Transition type.
	TransitionType PageTransitionType `json:"transitionType"`
}

type PageScreencastFrameMetadata struct {
	// Top offset in DIP.
	OffsetTop float64 `json:"offsetTop"`
	// Page scale factor.
	PageScaleFactor float64 `json:"pageScaleFactor"`
	// Device screen width in DIP.
	DeviceWidth float64 `json:"deviceWidth"`
	// Device screen height in DIP.
	DeviceHeight float64 `json:"deviceHeight"`
	// Position of horizontal scroll in CSS pixels.
	ScrollOffsetX float64 `json:"scrollOffsetX"`
	// Position of vertical scroll in CSS pixels.
	ScrollOffsetY float64 `json:"scrollOffsetY"`
	// Frame swap timestamp.
	Timestamp *NetworkTimeSinceEpoch `json:"timestamp,omitempty"`
}

type PageDialogType string

type PageAppManifestError struct {
	// Error message.
	Message string `json:"message"`
	// If critical, this is a non-recoverable parse error.
	Critical int `json:"critical"`
	// Error line.
	Line int `json:"line"`
	// Error column.
	Column int `json:"column"`
}

type PageAppManifestParsedProperties struct {
	// Computed scope value
	Scope string `json:"scope"`
}

type PageLayoutViewport struct {
	// Horizontal offset relative to the document (CSS pixels).
	PageX int `json:"pageX"`
	// Vertical offset relative to the document (CSS pixels).
	PageY int `json:"pageY"`
	// Width (CSS pixels), excludes scrollbar if present.
	ClientWidth int `json:"clientWidth"`
	// Height (CSS pixels), excludes scrollbar if present.
	ClientHeight int `json:"clientHeight"`
}

type PageVisualViewport struct {
	// Horizontal offset relative to the layout viewport (CSS pixels).
	OffsetX float64 `json:"offsetX"`
	// Vertical offset relative to the layout viewport (CSS pixels).
	OffsetY float64 `json:"offsetY"`
	// Horizontal offset relative to the document (CSS pixels).
	PageX float64 `json:"pageX"`
	// Vertical offset relative to the document (CSS pixels).
	PageY float64 `json:"pageY"`
	// Width (CSS pixels), excludes scrollbar if present.
	ClientWidth float64 `json:"clientWidth"`
	// Height (CSS pixels), excludes scrollbar if present.
	ClientHeight float64 `json:"clientHeight"`
	// Scale relative to the ideal viewport (size at width=device-width).
	Scale float64 `json:"scale"`
	// Page zoom factor (CSS to device independent pixels ratio).
	Zoom *float64 `json:"zoom,omitempty"`
}

type PageViewport struct {
	// X offset in device independent pixels (dip).
	X float64 `json:"x"`
	// Y offset in device independent pixels (dip).
	Y float64 `json:"y"`
	// Rectangle width in device independent pixels (dip).
	Width float64 `json:"width"`
	// Rectangle height in device independent pixels (dip).
	Height float64 `json:"height"`
	// Page scale factor.
	Scale float64 `json:"scale"`
}

type PageFontFamilies struct {
	// The standard font-family.
	Standard *string `json:"standard,omitempty"`
	// The fixed font-family.
	Fixed *string `json:"fixed,omitempty"`
	// The serif font-family.
	Serif *string `json:"serif,omitempty"`
	// The sansSerif font-family.
	SansSerif *string `json:"sansSerif,omitempty"`
	// The cursive font-family.
	Cursive *string `json:"cursive,omitempty"`
	// The fantasy font-family.
	Fantasy *string `json:"fantasy,omitempty"`
	// The math font-family.
	Math *string `json:"math,omitempty"`
}

type PageScriptFontFamilies struct {
	// Name of the script which these font families are defined for.
	Script string `json:"script"`
	// Generic font families collection for the script.
	FontFamilies PageFontFamilies `json:"fontFamilies"`
}

type PageFontSizes struct {
	// Default standard font size.
	Standard *int `json:"standard,omitempty"`
	// Default fixed font size.
	Fixed *int `json:"fixed,omitempty"`
}

type PageClientNavigationReason string

type PageClientNavigationDisposition string

type PageInstallabilityErrorArgument struct {
	// Argument name (e.g. name:'minimum-icon-size-in-pixels').
	Name string `json:"name"`
	// Argument value (e.g. value:'64').
	Value string `json:"value"`
}

type PageInstallabilityError struct {
	// The error id (e.g. 'manifest-missing-suitable-icon').
	ErrorID string `json:"errorId"`
	// The list of error arguments (e.g. {name:'minimum-icon-size-in-pixels', value:'64'}).
	ErrorArguments []PageInstallabilityErrorArgument `json:"errorArguments"`
}

type PageReferrerPolicy string

type PageCompilationCacheParams struct {
	// The URL of the script to produce a compilation cache entry for.
	URL string `json:"url"`
	// A hint to the backend whether eager compilation is recommended.
	Eager *bool `json:"eager,omitempty"`
}

type PageFileFilter struct {
	Name    *string  `json:"name,omitempty"`
	Accepts []string `json:"accepts,omitempty"`
}

type PageFileHandler struct {
	Action string              `json:"action"`
	Name   string              `json:"name"`
	Icons  []PageImageResource `json:"icons,omitempty"`
	// Mimic a map, name is the key, accepts is the value.
	Accepts []PageFileFilter `json:"accepts,omitempty"`
	// Won't repeat the enums, using string for easy comparison. Same as the
	LaunchType string `json:"launchType"`
}

type PageImageResource struct {
	// The src field in the definition, but changing to url in favor of
	URL   string  `json:"url"`
	Sizes *string `json:"sizes,omitempty"`
	Type  *string `json:"type,omitempty"`
}

type PageLaunchHandler struct {
	ClientMode string `json:"clientMode"`
}

type PageProtocolHandler struct {
	Protocol string `json:"protocol"`
	URL      string `json:"url"`
}

type PageRelatedApplication struct {
	ID  *string `json:"id,omitempty"`
	URL string  `json:"url"`
}

type PageScopeExtension struct {
	// Instead of using tuple, this field always returns the serialized string
	Origin            string `json:"origin"`
	HasOriginWildcard bool   `json:"hasOriginWildcard"`
}

type PageScreenshot struct {
	Image      PageImageResource `json:"image"`
	FormFactor string            `json:"formFactor"`
	Label      *string           `json:"label,omitempty"`
}

type PageShareTarget struct {
	Action  string `json:"action"`
	Method  string `json:"method"`
	Enctype string `json:"enctype"`
	// Embed the ShareTargetParams
	Title *string          `json:"title,omitempty"`
	Text  *string          `json:"text,omitempty"`
	URL   *string          `json:"url,omitempty"`
	Files []PageFileFilter `json:"files,omitempty"`
}

type PageShortcut struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type PageWebAppManifest struct {
	BackgroundColor *string `json:"backgroundColor,omitempty"`
	// The extra description provided by the manifest.
	Description *string `json:"description,omitempty"`
	Dir         *string `json:"dir,omitempty"`
	Display     *string `json:"display,omitempty"`
	// The overrided display mode controlled by the user.
	DisplayOverrides []string `json:"displayOverrides,omitempty"`
	// The handlers to open files.
	FileHandlers []PageFileHandler   `json:"fileHandlers,omitempty"`
	Icons        []PageImageResource `json:"icons,omitempty"`
	ID           *string             `json:"id,omitempty"`
	Lang         *string             `json:"lang,omitempty"`
	// TODO(crbug.com/1231886): This field is non-standard and part of a Chrome
	LaunchHandler             *PageLaunchHandler `json:"launchHandler,omitempty"`
	Name                      *string            `json:"name,omitempty"`
	Orientation               *string            `json:"orientation,omitempty"`
	PreferRelatedApplications *bool              `json:"preferRelatedApplications,omitempty"`
	// The handlers to open protocols.
	ProtocolHandlers    []PageProtocolHandler    `json:"protocolHandlers,omitempty"`
	RelatedApplications []PageRelatedApplication `json:"relatedApplications,omitempty"`
	Scope               *string                  `json:"scope,omitempty"`
	// Non-standard, see
	ScopeExtensions []PageScopeExtension `json:"scopeExtensions,omitempty"`
	// The screenshots used by chromium.
	Screenshots []PageScreenshot `json:"screenshots,omitempty"`
	ShareTarget *PageShareTarget `json:"shareTarget,omitempty"`
	ShortName   *string          `json:"shortName,omitempty"`
	Shortcuts   []PageShortcut   `json:"shortcuts,omitempty"`
	StartURL    *string          `json:"startUrl,omitempty"`
	ThemeColor  *string          `json:"themeColor,omitempty"`
}

type PageNavigationType string

type PageBackForwardCacheNotRestoredReason string

type PageBackForwardCacheNotRestoredReasonType string

type PageBackForwardCacheBlockingDetails struct {
	// Url of the file where blockage happened. Optional because of tests.
	URL *string `json:"url,omitempty"`
	// Function name where blockage happened. Optional because of anonymous functions and tests.
	Function *string `json:"function,omitempty"`
	// Line number in the script (0-based).
	LineNumber int `json:"lineNumber"`
	// Column number in the script (0-based).
	ColumnNumber int `json:"columnNumber"`
}

type PageBackForwardCacheNotRestoredExplanation struct {
	// Type of the reason
	Type PageBackForwardCacheNotRestoredReasonType `json:"type"`
	// Not restored reason
	Reason PageBackForwardCacheNotRestoredReason `json:"reason"`
	// Context associated with the reason. The meaning of this context is
	Context *string                               `json:"context,omitempty"`
	Details []PageBackForwardCacheBlockingDetails `json:"details,omitempty"`
}

type PageBackForwardCacheNotRestoredExplanationTree struct {
	// URL of each frame
	URL string `json:"url"`
	// Not restored reasons of each frame
	Explanations []PageBackForwardCacheNotRestoredExplanation `json:"explanations"`
	// Array of children frame
	Children []PageBackForwardCacheNotRestoredExplanationTree `json:"children"`
}

type PageAddScriptToEvaluateOnLoadParams struct {
	SessionID    string `json:"-"`
	ScriptSource string `json:"scriptSource"`
}

type PageAddScriptToEvaluateOnLoadResult struct {
	// Identifier of the added script.
	Identifier PageScriptIdentifier `json:"identifier"`
}

type PageAddScriptToEvaluateOnNewDocumentParams struct {
	SessionID string `json:"-"`
	Source    string `json:"source"`
	// If specified, creates an isolated world with the given name and evaluates given script in it.
	WorldName *string `json:"worldName,omitempty"`
	// Specifies whether command line API should be available to the script, defaults
	IncludeCommandLineAPI *bool `json:"includeCommandLineAPI,omitempty"`
	// If true, runs the script immediately on existing execution contexts or worlds.
	RunImmediately *bool `json:"runImmediately,omitempty"`
}

type PageAddScriptToEvaluateOnNewDocumentResult struct {
	// Identifier of the added script.
	Identifier PageScriptIdentifier `json:"identifier"`
}

type PageBringToFrontParams struct {
	SessionID string `json:"-"`
}

type PageBringToFrontResult struct {
}

type PageCaptureScreenshotParams struct {
	SessionID string `json:"-"`
	// Image compression format (defaults to png).
	Format *string `json:"format,omitempty"`
	// Compression quality from range [0..100] (jpeg only).
	Quality *int `json:"quality,omitempty"`
	// Capture the screenshot of a given region only.
	Clip *PageViewport `json:"clip,omitempty"`
	// Capture the screenshot from the surface, rather than the view. Defaults to true.
	FromSurface *bool `json:"fromSurface,omitempty"`
	// Capture the screenshot beyond the viewport. Defaults to false.
	CaptureBeyondViewport *bool `json:"captureBeyondViewport,omitempty"`
	// Optimize image encoding for speed, not for resulting size (defaults to false)
	OptimizeForSpeed *bool `json:"optimizeForSpeed,omitempty"`
}

type PageCaptureScreenshotResult struct {
	// Base64-encoded image data. (Encoded as a base64 string when passed over JSON)
	Data string `json:"data"`
}

type PageCaptureSnapshotParams struct {
	SessionID string `json:"-"`
	// Format (defaults to mhtml).
	Format *string `json:"format,omitempty"`
}

type PageCaptureSnapshotResult struct {
	// Serialized page data.
	Data string `json:"data"`
}

type PageClearDeviceMetricsOverrideParams struct {
	SessionID string `json:"-"`
}

type PageClearDeviceMetricsOverrideResult struct {
}

type PageClearDeviceOrientationOverrideParams struct {
	SessionID string `json:"-"`
}

type PageClearDeviceOrientationOverrideResult struct {
}

type PageClearGeolocationOverrideParams struct {
	SessionID string `json:"-"`
}

type PageClearGeolocationOverrideResult struct {
}

type PageCreateIsolatedWorldParams struct {
	SessionID string `json:"-"`
	// Id of the frame in which the isolated world should be created.
	FrameID PageFrameID `json:"frameId"`
	// An optional name which is reported in the Execution Context.
	WorldName *string `json:"worldName,omitempty"`
	// Whether or not universal access should be granted to the isolated world. This is a powerful
	GrantUniveralAccess *bool `json:"grantUniveralAccess,omitempty"`
}

type PageCreateIsolatedWorldResult struct {
	// Execution context of the isolated world.
	ExecutionContextID RuntimeExecutionContextID `json:"executionContextId"`
}

type PageDeleteCookieParams struct {
	SessionID string `json:"-"`
	// Name of the cookie to remove.
	CookieName string `json:"cookieName"`
	// URL to match cooke domain and path.
	URL string `json:"url"`
}

type PageDeleteCookieResult struct {
}

type PageDisableParams struct {
	SessionID string `json:"-"`
}

type PageDisableResult struct {
}

type PageEnableParams struct {
	SessionID string `json:"-"`
	// If true, the `Page.fileChooserOpened` event will be emitted regardless of the state set by
	EnableFileChooserOpenedEvent *bool `json:"enableFileChooserOpenedEvent,omitempty"`
}

type PageEnableResult struct {
}

type PageGetAppManifestParams struct {
	SessionID  string  `json:"-"`
	ManifestID *string `json:"manifestId,omitempty"`
}

type PageGetAppManifestResult struct {
	// Manifest location.
	URL    string                 `json:"url"`
	Errors []PageAppManifestError `json:"errors"`
	// Manifest content.
	Data *string `json:"data,omitempty"`
	// Parsed manifest properties. Deprecated, use manifest instead.
	Parsed   *PageAppManifestParsedProperties `json:"parsed,omitempty"`
	Manifest PageWebAppManifest               `json:"manifest"`
}

type PageGetInstallabilityErrorsParams struct {
	SessionID string `json:"-"`
}

type PageGetInstallabilityErrorsResult struct {
	InstallabilityErrors []PageInstallabilityError `json:"installabilityErrors"`
}

type PageGetManifestIconsParams struct {
	SessionID string `json:"-"`
}

type PageGetManifestIconsResult struct {
	PrimaryIcon *string `json:"primaryIcon,omitempty"`
}

type PageGetAppIDParams struct {
	SessionID string `json:"-"`
}

type PageGetAppIDResult struct {
	// App id, either from manifest's id attribute or computed from start_url
	AppID *string `json:"appId,omitempty"`
	// Recommendation for manifest's id attribute to match current id computed from start_url
	RecommendedID *string `json:"recommendedId,omitempty"`
}

type PageGetAdScriptAncestryParams struct {
	SessionID string      `json:"-"`
	FrameID   PageFrameID `json:"frameId"`
}

type PageGetAdScriptAncestryResult struct {
	// The ancestry chain of ad script identifiers leading to this frame's
	AdScriptAncestry *PageAdScriptAncestry `json:"adScriptAncestry,omitempty"`
}

type PageGetFrameTreeParams struct {
	SessionID string `json:"-"`
}

type PageGetFrameTreeResult struct {
	// Present frame tree structure.
	FrameTree PageFrameTree `json:"frameTree"`
}

type PageGetLayoutMetricsParams struct {
	SessionID string `json:"-"`
}

type PageGetLayoutMetricsResult struct {
	// Deprecated metrics relating to the layout viewport. Is in device pixels. Use `cssLayoutViewport` instead.
	LayoutViewport PageLayoutViewport `json:"layoutViewport"`
	// Deprecated metrics relating to the visual viewport. Is in device pixels. Use `cssVisualViewport` instead.
	VisualViewport PageVisualViewport `json:"visualViewport"`
	// Deprecated size of scrollable area. Is in DP. Use `cssContentSize` instead.
	ContentSize DOMRect `json:"contentSize"`
	// Metrics relating to the layout viewport in CSS pixels.
	CSSLayoutViewport PageLayoutViewport `json:"cssLayoutViewport"`
	// Metrics relating to the visual viewport in CSS pixels.
	CSSVisualViewport PageVisualViewport `json:"cssVisualViewport"`
	// Size of scrollable area in CSS pixels.
	CSSContentSize DOMRect `json:"cssContentSize"`
}

type PageGetNavigationHistoryParams struct {
	SessionID string `json:"-"`
}

type PageGetNavigationHistoryResult struct {
	// Index of the current navigation history entry.
	CurrentIndex int `json:"currentIndex"`
	// Array of navigation history entries.
	Entries []PageNavigationEntry `json:"entries"`
}

type PageResetNavigationHistoryParams struct {
	SessionID string `json:"-"`
}

type PageResetNavigationHistoryResult struct {
}

type PageGetResourceContentParams struct {
	SessionID string `json:"-"`
	// Frame id to get resource for.
	FrameID PageFrameID `json:"frameId"`
	// URL of the resource to get content for.
	URL string `json:"url"`
}

type PageGetResourceContentResult struct {
	// Resource content.
	Content string `json:"content"`
	// True, if content was served as base64.
	Base64Encoded bool `json:"base64Encoded"`
}

type PageGetResourceTreeParams struct {
	SessionID string `json:"-"`
}

type PageGetResourceTreeResult struct {
	// Present frame / resource tree structure.
	FrameTree PageFrameResourceTree `json:"frameTree"`
}

type PageHandleJavaScriptDialogParams struct {
	SessionID string `json:"-"`
	// Whether to accept or dismiss the dialog.
	Accept bool `json:"accept"`
	// The text to enter into the dialog prompt before accepting. Used only if this is a prompt
	PromptText *string `json:"promptText,omitempty"`
}

type PageHandleJavaScriptDialogResult struct {
}

type PageNavigateParams struct {
	SessionID string `json:"-"`
	// URL to navigate the page to.
	URL string `json:"url"`
	// Referrer URL.
	Referrer *string `json:"referrer,omitempty"`
	// Intended transition type.
	TransitionType *PageTransitionType `json:"transitionType,omitempty"`
	// Frame id to navigate, if not specified navigates the top frame.
	FrameID *PageFrameID `json:"frameId,omitempty"`
	// Referrer-policy used for the navigation.
	ReferrerPolicy *PageReferrerPolicy `json:"referrerPolicy,omitempty"`
}

type PageNavigateResult struct {
	// Frame id that has navigated (or failed to navigate)
	FrameID PageFrameID `json:"frameId"`
	// Loader identifier. This is omitted in case of same-document navigation,
	LoaderID *NetworkLoaderID `json:"loaderId,omitempty"`
	// User friendly error message, present if and only if navigation has failed.
	ErrorText *string `json:"errorText,omitempty"`
	// Whether the navigation resulted in a download.
	IsDownload *bool `json:"isDownload,omitempty"`
}

type PageNavigateToHistoryEntryParams struct {
	SessionID string `json:"-"`
	// Unique id of the entry to navigate to.
	EntryID int `json:"entryId"`
}

type PageNavigateToHistoryEntryResult struct {
}

type PagePrintToPDFParams struct {
	SessionID string `json:"-"`
	// Paper orientation. Defaults to false.
	Landscape *bool `json:"landscape,omitempty"`
	// Display header and footer. Defaults to false.
	DisplayHeaderFooter *bool `json:"displayHeaderFooter,omitempty"`
	// Print background graphics. Defaults to false.
	PrintBackground *bool `json:"printBackground,omitempty"`
	// Scale of the webpage rendering. Defaults to 1.
	Scale *float64 `json:"scale,omitempty"`
	// Paper width in inches. Defaults to 8.5 inches.
	PaperWidth *float64 `json:"paperWidth,omitempty"`
	// Paper height in inches. Defaults to 11 inches.
	PaperHeight *float64 `json:"paperHeight,omitempty"`
	// Top margin in inches. Defaults to 1cm (~0.4 inches).
	MarginTop *float64 `json:"marginTop,omitempty"`
	// Bottom margin in inches. Defaults to 1cm (~0.4 inches).
	MarginBottom *float64 `json:"marginBottom,omitempty"`
	// Left margin in inches. Defaults to 1cm (~0.4 inches).
	MarginLeft *float64 `json:"marginLeft,omitempty"`
	// Right margin in inches. Defaults to 1cm (~0.4 inches).
	MarginRight *float64 `json:"marginRight,omitempty"`
	// Paper ranges to print, one based, e.g., '1-5, 8, 11-13'. Pages are
	PageRanges *string `json:"pageRanges,omitempty"`
	// HTML template for the print header. Should be valid HTML markup with following
	HeaderTemplate *string `json:"headerTemplate,omitempty"`
	// HTML template for the print footer. Should use the same format as the `headerTemplate`.
	FooterTemplate *string `json:"footerTemplate,omitempty"`
	// Whether or not to prefer page size as defined by css. Defaults to false,
	PreferCSSPageSize *bool `json:"preferCSSPageSize,omitempty"`
	// return as stream
	TransferMode *string `json:"transferMode,omitempty"`
	// Whether or not to generate tagged (accessible) PDF. Defaults to embedder choice.
	GenerateTaggedPDF *bool `json:"generateTaggedPDF,omitempty"`
	// Whether or not to embed the document outline into the PDF.
	GenerateDocumentOutline *bool `json:"generateDocumentOutline,omitempty"`
}

type PagePrintToPDFResult struct {
	// Base64-encoded pdf data. Empty if |returnAsStream| is specified. (Encoded as a base64 string when passed over JSON)
	Data string `json:"data"`
	// A handle of the stream that holds resulting PDF data.
	Stream *IOStreamHandle `json:"stream,omitempty"`
}

type PageReloadParams struct {
	SessionID string `json:"-"`
	// If true, browser cache is ignored (as if the user pressed Shift+refresh).
	IgnoreCache *bool `json:"ignoreCache,omitempty"`
	// If set, the script will be injected into all frames of the inspected page after reload.
	ScriptToEvaluateOnLoad *string `json:"scriptToEvaluateOnLoad,omitempty"`
	// If set, an error will be thrown if the target page's main frame's
	LoaderID *NetworkLoaderID `json:"loaderId,omitempty"`
}

type PageReloadResult struct {
}

type PageRemoveScriptToEvaluateOnLoadParams struct {
	SessionID  string               `json:"-"`
	Identifier PageScriptIdentifier `json:"identifier"`
}

type PageRemoveScriptToEvaluateOnLoadResult struct {
}

type PageRemoveScriptToEvaluateOnNewDocumentParams struct {
	SessionID  string               `json:"-"`
	Identifier PageScriptIdentifier `json:"identifier"`
}

type PageRemoveScriptToEvaluateOnNewDocumentResult struct {
}

type PageScreencastFrameAckParams struct {
	SessionID string `json:"-"`
	// Frame number.
	SessionIDValue int `json:"sessionId"`
}

type PageScreencastFrameAckResult struct {
}

type PageSearchInResourceParams struct {
	SessionID string `json:"-"`
	// Frame id for resource to search in.
	FrameID PageFrameID `json:"frameId"`
	// URL of the resource to search in.
	URL string `json:"url"`
	// String to search for.
	Query string `json:"query"`
	// If true, search is case sensitive.
	CaseSensitive *bool `json:"caseSensitive,omitempty"`
	// If true, treats string parameter as regex.
	IsRegex *bool `json:"isRegex,omitempty"`
}

type PageSearchInResourceResult struct {
	// List of search matches.
	Result []DebuggerSearchMatch `json:"result"`
}

type PageSetAdBlockingEnabledParams struct {
	SessionID string `json:"-"`
	// Whether to block ads.
	Enabled bool `json:"enabled"`
}

type PageSetAdBlockingEnabledResult struct {
}

type PageSetBypassCSPParams struct {
	SessionID string `json:"-"`
	// Whether to bypass page CSP.
	Enabled bool `json:"enabled"`
}

type PageSetBypassCSPResult struct {
}

type PageGetPermissionsPolicyStateParams struct {
	SessionID string      `json:"-"`
	FrameID   PageFrameID `json:"frameId"`
}

type PageGetPermissionsPolicyStateResult struct {
	States []PagePermissionsPolicyFeatureState `json:"states"`
}

type PageGetOriginTrialsParams struct {
	SessionID string      `json:"-"`
	FrameID   PageFrameID `json:"frameId"`
}

type PageGetOriginTrialsResult struct {
	OriginTrials []PageOriginTrial `json:"originTrials"`
}

type PageSetDeviceMetricsOverrideParams struct {
	SessionID string `json:"-"`
	// Overriding width value in pixels (minimum 0, maximum 10000000). 0 disables the override.
	Width int `json:"width"`
	// Overriding height value in pixels (minimum 0, maximum 10000000). 0 disables the override.
	Height int `json:"height"`
	// Overriding device scale factor value. 0 disables the override.
	DeviceScaleFactor float64 `json:"deviceScaleFactor"`
	// Whether to emulate mobile device. This includes viewport meta tag, overlay scrollbars, text
	Mobile bool `json:"mobile"`
	// Scale to apply to resulting view image.
	Scale *float64 `json:"scale,omitempty"`
	// Overriding screen width value in pixels (minimum 0, maximum 10000000).
	ScreenWidth *int `json:"screenWidth,omitempty"`
	// Overriding screen height value in pixels (minimum 0, maximum 10000000).
	ScreenHeight *int `json:"screenHeight,omitempty"`
	// Overriding view X position on screen in pixels (minimum 0, maximum 10000000).
	PositionX *int `json:"positionX,omitempty"`
	// Overriding view Y position on screen in pixels (minimum 0, maximum 10000000).
	PositionY *int `json:"positionY,omitempty"`
	// Do not set visible view size, rely upon explicit setVisibleSize call.
	DontSetVisibleSize *bool `json:"dontSetVisibleSize,omitempty"`
	// Screen orientation override.
	ScreenOrientation *EmulationScreenOrientation `json:"screenOrientation,omitempty"`
	// The viewport dimensions and scale. If not set, the override is cleared.
	Viewport *PageViewport `json:"viewport,omitempty"`
}

type PageSetDeviceMetricsOverrideResult struct {
}

type PageSetDeviceOrientationOverrideParams struct {
	SessionID string `json:"-"`
	// Mock alpha
	Alpha float64 `json:"alpha"`
	// Mock beta
	Beta float64 `json:"beta"`
	// Mock gamma
	Gamma float64 `json:"gamma"`
}

type PageSetDeviceOrientationOverrideResult struct {
}

type PageSetFontFamiliesParams struct {
	SessionID string `json:"-"`
	// Specifies font families to set. If a font family is not specified, it won't be changed.
	FontFamilies PageFontFamilies `json:"fontFamilies"`
	// Specifies font families to set for individual scripts.
	ForScripts []PageScriptFontFamilies `json:"forScripts,omitempty"`
}

type PageSetFontFamiliesResult struct {
}

type PageSetFontSizesParams struct {
	SessionID string `json:"-"`
	// Specifies font sizes to set. If a font size is not specified, it won't be changed.
	FontSizes PageFontSizes `json:"fontSizes"`
}

type PageSetFontSizesResult struct {
}

type PageSetDocumentContentParams struct {
	SessionID string `json:"-"`
	// Frame id to set HTML for.
	FrameID PageFrameID `json:"frameId"`
	// HTML content to set.
	HTML string `json:"html"`
}

type PageSetDocumentContentResult struct {
}

type PageSetDownloadBehaviorParams struct {
	SessionID string `json:"-"`
	// Whether to allow all or deny all download requests, or use default Chrome behavior if
	Behavior string `json:"behavior"`
	// The default path to save downloaded files to. This is required if behavior is set to 'allow'
	DownloadPath *string `json:"downloadPath,omitempty"`
}

type PageSetDownloadBehaviorResult struct {
}

type PageSetGeolocationOverrideParams struct {
	SessionID string `json:"-"`
	// Mock latitude
	Latitude *float64 `json:"latitude,omitempty"`
	// Mock longitude
	Longitude *float64 `json:"longitude,omitempty"`
	// Mock accuracy
	Accuracy *float64 `json:"accuracy,omitempty"`
}

type PageSetGeolocationOverrideResult struct {
}

type PageSetLifecycleEventsEnabledParams struct {
	SessionID string `json:"-"`
	// If true, starts emitting lifecycle events.
	Enabled bool `json:"enabled"`
}

type PageSetLifecycleEventsEnabledResult struct {
}

type PageSetTouchEmulationEnabledParams struct {
	SessionID string `json:"-"`
	// Whether the touch event emulation should be enabled.
	Enabled bool `json:"enabled"`
	// Touch/gesture events configuration. Default: current platform.
	Configuration *string `json:"configuration,omitempty"`
}

type PageSetTouchEmulationEnabledResult struct {
}

type PageStartScreencastParams struct {
	SessionID string `json:"-"`
	// Image compression format.
	Format *string `json:"format,omitempty"`
	// Compression quality from range [0..100].
	Quality *int `json:"quality,omitempty"`
	// Maximum screenshot width.
	MaxWidth *int `json:"maxWidth,omitempty"`
	// Maximum screenshot height.
	MaxHeight *int `json:"maxHeight,omitempty"`
	// Send every n-th frame.
	EveryNthFrame *int `json:"everyNthFrame,omitempty"`
}

type PageStartScreencastResult struct {
}

type PageStopLoadingParams struct {
	SessionID string `json:"-"`
}

type PageStopLoadingResult struct {
}

type PageCrashParams struct {
	SessionID string `json:"-"`
}

type PageCrashResult struct {
}

type PageCloseParams struct {
	SessionID string `json:"-"`
}

type PageCloseResult struct {
}

type PageSetWebLifecycleStateParams struct {
	SessionID string `json:"-"`
	// Target lifecycle state
	State string `json:"state"`
}

type PageSetWebLifecycleStateResult struct {
}

type PageStopScreencastParams struct {
	SessionID string `json:"-"`
}

type PageStopScreencastResult struct {
}

type PageProduceCompilationCacheParams struct {
	SessionID string                       `json:"-"`
	Scripts   []PageCompilationCacheParams `json:"scripts"`
}

type PageProduceCompilationCacheResult struct {
}

type PageAddCompilationCacheParams struct {
	SessionID string `json:"-"`
	URL       string `json:"url"`
	// Base64-encoded data (Encoded as a base64 string when passed over JSON)
	Data string `json:"data"`
}

type PageAddCompilationCacheResult struct {
}

type PageClearCompilationCacheParams struct {
	SessionID string `json:"-"`
}

type PageClearCompilationCacheResult struct {
}

type PageSetSPCTransactionModeParams struct {
	SessionID string `json:"-"`
	Mode      string `json:"mode"`
}

type PageSetSPCTransactionModeResult struct {
}

type PageSetRPHRegistrationModeParams struct {
	SessionID string `json:"-"`
	Mode      string `json:"mode"`
}

type PageSetRPHRegistrationModeResult struct {
}

type PageGenerateTestReportParams struct {
	SessionID string `json:"-"`
	// Message to be displayed in the report.
	Message string `json:"message"`
	// Specifies the endpoint group to deliver the report to.
	Group *string `json:"group,omitempty"`
}

type PageGenerateTestReportResult struct {
}

type PageWaitForDebuggerParams struct {
	SessionID string `json:"-"`
}

type PageWaitForDebuggerResult struct {
}

type PageSetInterceptFileChooserDialogParams struct {
	SessionID string `json:"-"`
	Enabled   bool   `json:"enabled"`
	// If true, cancels the dialog by emitting relevant events (if any)
	Cancel *bool `json:"cancel,omitempty"`
}

type PageSetInterceptFileChooserDialogResult struct {
}

type PageSetPrerenderingAllowedParams struct {
	SessionID string `json:"-"`
	IsAllowed bool   `json:"isAllowed"`
}

type PageSetPrerenderingAllowedResult struct {
}

type PageGetAnnotatedPageContentParams struct {
	SessionID string `json:"-"`
	// Whether to include actionable information. Defaults to true.
	IncludeActionableInformation *bool `json:"includeActionableInformation,omitempty"`
}

type PageGetAnnotatedPageContentResult struct {
	// The annotated page content as a base64 encoded protobuf.
	Content string `json:"content"`
}

type PageDOMContentEventFiredEvent struct {
	Timestamp NetworkMonotonicTime `json:"timestamp"`
}

type PageFileChooserOpenedEvent struct {
	// Id of the frame containing input node.
	FrameID PageFrameID `json:"frameId"`
	// Input mode.
	Mode string `json:"mode"`
	// Input node id. Only present for file choosers opened via an `<input type="file">` element.
	BackendNodeID *DOMBackendNodeID `json:"backendNodeId,omitempty"`
}

type PageFrameAttachedEvent struct {
	// Id of the frame that has been attached.
	FrameID PageFrameID `json:"frameId"`
	// Parent frame identifier.
	ParentFrameID PageFrameID `json:"parentFrameId"`
	// JavaScript stack trace of when frame was attached, only set if frame initiated from script.
	Stack *RuntimeStackTrace `json:"stack,omitempty"`
}

type PageFrameClearedScheduledNavigationEvent struct {
	// Id of the frame that has cleared its scheduled navigation.
	FrameID PageFrameID `json:"frameId"`
}

type PageFrameDetachedEvent struct {
	// Id of the frame that has been detached.
	FrameID PageFrameID `json:"frameId"`
	Reason  string      `json:"reason"`
}

type PageFrameSubtreeWillBeDetachedEvent struct {
	// Id of the frame that is the root of the subtree that will be detached.
	FrameID PageFrameID `json:"frameId"`
}

type PageFrameNavigatedEvent struct {
	// Frame object.
	Frame PageFrame          `json:"frame"`
	Type  PageNavigationType `json:"type"`
}

type PageDocumentOpenedEvent struct {
	// Frame object.
	Frame PageFrame `json:"frame"`
}

type PageFrameResizedEvent struct {
}

type PageFrameStartedNavigatingEvent struct {
	// ID of the frame that is being navigated.
	FrameID PageFrameID `json:"frameId"`
	// The URL the navigation started with. The final URL can be different.
	URL string `json:"url"`
	// Loader identifier. Even though it is present in case of same-document
	LoaderID       NetworkLoaderID `json:"loaderId"`
	NavigationType string          `json:"navigationType"`
}

type PageFrameRequestedNavigationEvent struct {
	// Id of the frame that is being navigated.
	FrameID PageFrameID `json:"frameId"`
	// The reason for the navigation.
	Reason PageClientNavigationReason `json:"reason"`
	// The destination URL for the requested navigation.
	URL string `json:"url"`
	// The disposition for the navigation.
	Disposition PageClientNavigationDisposition `json:"disposition"`
}

type PageFrameScheduledNavigationEvent struct {
	// Id of the frame that has scheduled a navigation.
	FrameID PageFrameID `json:"frameId"`
	// Delay (in seconds) until the navigation is scheduled to begin. The navigation is not
	Delay float64 `json:"delay"`
	// The reason for the navigation.
	Reason PageClientNavigationReason `json:"reason"`
	// The destination URL for the scheduled navigation.
	URL string `json:"url"`
}

type PageFrameStartedLoadingEvent struct {
	// Id of the frame that has started loading.
	FrameID PageFrameID `json:"frameId"`
}

type PageFrameStoppedLoadingEvent struct {
	// Id of the frame that has stopped loading.
	FrameID PageFrameID `json:"frameId"`
}

type PageDownloadWillBeginEvent struct {
	// Id of the frame that caused download to begin.
	FrameID PageFrameID `json:"frameId"`
	// Global unique identifier of the download.
	Guid string `json:"guid"`
	// URL of the resource being downloaded.
	URL string `json:"url"`
	// Suggested file name of the resource (the actual name of the file saved on disk may differ).
	SuggestedFilename string `json:"suggestedFilename"`
}

type PageDownloadProgressEvent struct {
	// Global unique identifier of the download.
	Guid string `json:"guid"`
	// Total expected bytes to download.
	TotalBytes float64 `json:"totalBytes"`
	// Total bytes received.
	ReceivedBytes float64 `json:"receivedBytes"`
	// Download status.
	State string `json:"state"`
}

type PageInterstitialHiddenEvent struct {
}

type PageInterstitialShownEvent struct {
}

type PageJavascriptDialogClosedEvent struct {
	// Frame id.
	FrameID PageFrameID `json:"frameId"`
	// Whether dialog was confirmed.
	Result bool `json:"result"`
	// User input in case of prompt.
	UserInput string `json:"userInput"`
}

type PageJavascriptDialogOpeningEvent struct {
	// Frame url.
	URL string `json:"url"`
	// Frame id.
	FrameID PageFrameID `json:"frameId"`
	// Message that will be displayed by the dialog.
	Message string `json:"message"`
	// Dialog type.
	Type PageDialogType `json:"type"`
	// True iff browser is capable showing or acting on the given dialog. When browser has no
	HasBrowserHandler bool `json:"hasBrowserHandler"`
	// Default dialog prompt.
	DefaultPrompt *string `json:"defaultPrompt,omitempty"`
}

type PageLifecycleEventEvent struct {
	// Id of the frame.
	FrameID PageFrameID `json:"frameId"`
	// Loader identifier. Empty string if the request is fetched from worker.
	LoaderID  NetworkLoaderID      `json:"loaderId"`
	Name      string               `json:"name"`
	Timestamp NetworkMonotonicTime `json:"timestamp"`
}

type PageBackForwardCacheNotUsedEvent struct {
	// The loader id for the associated navigation.
	LoaderID NetworkLoaderID `json:"loaderId"`
	// The frame id of the associated frame.
	FrameID PageFrameID `json:"frameId"`
	// Array of reasons why the page could not be cached. This must not be empty.
	NotRestoredExplanations []PageBackForwardCacheNotRestoredExplanation `json:"notRestoredExplanations"`
	// Tree structure of reasons why the page could not be cached for each frame.
	NotRestoredExplanationsTree *PageBackForwardCacheNotRestoredExplanationTree `json:"notRestoredExplanationsTree,omitempty"`
}

type PageLoadEventFiredEvent struct {
	Timestamp NetworkMonotonicTime `json:"timestamp"`
}

type PageNavigatedWithinDocumentEvent struct {
	// Id of the frame.
	FrameID PageFrameID `json:"frameId"`
	// Frame's new url.
	URL string `json:"url"`
	// Navigation type
	NavigationType string `json:"navigationType"`
}

type PageScreencastFrameEvent struct {
	// Base64-encoded compressed image. (Encoded as a base64 string when passed over JSON)
	Data string `json:"data"`
	// Screencast frame metadata.
	Metadata PageScreencastFrameMetadata `json:"metadata"`
	// Frame number.
	SessionID int `json:"sessionId"`
}

type PageScreencastVisibilityChangedEvent struct {
	// True if the page is visible.
	Visible bool `json:"visible"`
}

type PageWindowOpenEvent struct {
	// The URL for the new window.
	URL string `json:"url"`
	// Window name.
	WindowName string `json:"windowName"`
	// An array of enabled window features.
	WindowFeatures []string `json:"windowFeatures"`
	// Whether or not it was triggered by user gesture.
	UserGesture bool `json:"userGesture"`
}

type PageCompilationCacheProducedEvent struct {
	URL string `json:"url"`
	// Base64-encoded data (Encoded as a base64 string when passed over JSON)
	Data string `json:"data"`
}

type PerformanceMetric struct {
	// Metric name.
	Name string `json:"name"`
	// Metric value.
	Value float64 `json:"value"`
}

type PerformanceDisableParams struct {
	SessionID string `json:"-"`
}

type PerformanceDisableResult struct {
}

type PerformanceEnableParams struct {
	SessionID string `json:"-"`
	// Time domain to use for collecting and reporting duration metrics.
	TimeDomain *string `json:"timeDomain,omitempty"`
}

type PerformanceEnableResult struct {
}

type PerformanceSetTimeDomainParams struct {
	SessionID string `json:"-"`
	// Time domain
	TimeDomain string `json:"timeDomain"`
}

type PerformanceSetTimeDomainResult struct {
}

type PerformanceGetMetricsParams struct {
	SessionID string `json:"-"`
}

type PerformanceGetMetricsResult struct {
	// Current values for run-time metrics.
	Metrics []PerformanceMetric `json:"metrics"`
}

type PerformanceMetricsEvent struct {
	// Current values of the metrics.
	Metrics []PerformanceMetric `json:"metrics"`
	// Timestamp title.
	Title string `json:"title"`
}

type PerformanceTimelineLargestContentfulPaint struct {
	RenderTime NetworkTimeSinceEpoch `json:"renderTime"`
	LoadTime   NetworkTimeSinceEpoch `json:"loadTime"`
	// The number of pixels being painted.
	Size float64 `json:"size"`
	// The id attribute of the element, if available.
	ElementID *string `json:"elementId,omitempty"`
	// The URL of the image (may be trimmed).
	URL    *string           `json:"url,omitempty"`
	NodeID *DOMBackendNodeID `json:"nodeId,omitempty"`
}

type PerformanceTimelineLayoutShiftAttribution struct {
	PreviousRect DOMRect           `json:"previousRect"`
	CurrentRect  DOMRect           `json:"currentRect"`
	NodeID       *DOMBackendNodeID `json:"nodeId,omitempty"`
}

type PerformanceTimelineLayoutShift struct {
	// Score increment produced by this event.
	Value          float64                                     `json:"value"`
	HadRecentInput bool                                        `json:"hadRecentInput"`
	LastInputTime  NetworkTimeSinceEpoch                       `json:"lastInputTime"`
	Sources        []PerformanceTimelineLayoutShiftAttribution `json:"sources"`
}

type PerformanceTimelineTimelineEvent struct {
	// Identifies the frame that this event is related to. Empty for non-frame targets.
	FrameID PageFrameID `json:"frameId"`
	// The event type, as specified in https://w3c.github.io/performance-timeline/#dom-performanceentry-entrytype
	Type string `json:"type"`
	// Name may be empty depending on the type.
	Name string `json:"name"`
	// Time in seconds since Epoch, monotonically increasing within document lifetime.
	Time NetworkTimeSinceEpoch `json:"time"`
	// Event duration, if applicable.
	Duration           *float64                                   `json:"duration,omitempty"`
	LcpDetails         *PerformanceTimelineLargestContentfulPaint `json:"lcpDetails,omitempty"`
	LayoutShiftDetails *PerformanceTimelineLayoutShift            `json:"layoutShiftDetails,omitempty"`
}

type PerformanceTimelineEnableParams struct {
	SessionID string `json:"-"`
	// The types of event to report, as specified in
	EventTypes []string `json:"eventTypes"`
}

type PerformanceTimelineEnableResult struct {
}

type PerformanceTimelineTimelineEventAddedEvent struct {
	Event PerformanceTimelineTimelineEvent `json:"event"`
}

type PreloadRuleSetID string

type PreloadRuleSet struct {
	ID PreloadRuleSetID `json:"id"`
	// Identifies a document which the rule set is associated with.
	LoaderID NetworkLoaderID `json:"loaderId"`
	// Source text of JSON representing the rule set. If it comes from
	SourceText string `json:"sourceText"`
	// A speculation rule set is either added through an inline
	BackendNodeID *DOMBackendNodeID `json:"backendNodeId,omitempty"`
	URL           *string           `json:"url,omitempty"`
	RequestID     *NetworkRequestID `json:"requestId,omitempty"`
	// Error information
	ErrorType *PreloadRuleSetErrorType `json:"errorType,omitempty"`
	// TODO(https://crbug.com/1425354): Replace this property with structured error.
	ErrorMessage *string `json:"errorMessage,omitempty"`
	// For more details, see:
	Tag *string `json:"tag,omitempty"`
}

type PreloadRuleSetErrorType string

type PreloadSpeculationAction string

type PreloadSpeculationTargetHint string

type PreloadPreloadingAttemptKey struct {
	LoaderID   NetworkLoaderID               `json:"loaderId"`
	Action     PreloadSpeculationAction      `json:"action"`
	URL        string                        `json:"url"`
	TargetHint *PreloadSpeculationTargetHint `json:"targetHint,omitempty"`
}

type PreloadPreloadingAttemptSource struct {
	Key        PreloadPreloadingAttemptKey `json:"key"`
	RuleSetIds []PreloadRuleSetID          `json:"ruleSetIds"`
	NodeIds    []DOMBackendNodeID          `json:"nodeIds"`
}

type PreloadPreloadPipelineID string

type PreloadPrerenderFinalStatus string

type PreloadPreloadingStatus string

type PreloadPrefetchStatus string

type PreloadPrerenderMismatchedHeaders struct {
	HeaderName      string  `json:"headerName"`
	InitialValue    *string `json:"initialValue,omitempty"`
	ActivationValue *string `json:"activationValue,omitempty"`
}

type PreloadEnableParams struct {
	SessionID string `json:"-"`
}

type PreloadEnableResult struct {
}

type PreloadDisableParams struct {
	SessionID string `json:"-"`
}

type PreloadDisableResult struct {
}

type PreloadRuleSetUpdatedEvent struct {
	RuleSet PreloadRuleSet `json:"ruleSet"`
}

type PreloadRuleSetRemovedEvent struct {
	ID PreloadRuleSetID `json:"id"`
}

type PreloadPreloadEnabledStateUpdatedEvent struct {
	DisabledByPreference                        bool `json:"disabledByPreference"`
	DisabledByDataSaver                         bool `json:"disabledByDataSaver"`
	DisabledByBatterySaver                      bool `json:"disabledByBatterySaver"`
	DisabledByHoldbackPrefetchSpeculationRules  bool `json:"disabledByHoldbackPrefetchSpeculationRules"`
	DisabledByHoldbackPrerenderSpeculationRules bool `json:"disabledByHoldbackPrerenderSpeculationRules"`
}

type PreloadPrefetchStatusUpdatedEvent struct {
	Key        PreloadPreloadingAttemptKey `json:"key"`
	PipelineID PreloadPreloadPipelineID    `json:"pipelineId"`
	// The frame id of the frame initiating prefetch.
	InitiatingFrameID PageFrameID             `json:"initiatingFrameId"`
	PrefetchURL       string                  `json:"prefetchUrl"`
	Status            PreloadPreloadingStatus `json:"status"`
	PrefetchStatus    PreloadPrefetchStatus   `json:"prefetchStatus"`
	RequestID         NetworkRequestID        `json:"requestId"`
}

type PreloadPrerenderStatusUpdatedEvent struct {
	Key             PreloadPreloadingAttemptKey  `json:"key"`
	PipelineID      PreloadPreloadPipelineID     `json:"pipelineId"`
	Status          PreloadPreloadingStatus      `json:"status"`
	PrerenderStatus *PreloadPrerenderFinalStatus `json:"prerenderStatus,omitempty"`
	// This is used to give users more information about the name of Mojo interface
	DisallowedMojoInterface *string                             `json:"disallowedMojoInterface,omitempty"`
	MismatchedHeaders       []PreloadPrerenderMismatchedHeaders `json:"mismatchedHeaders,omitempty"`
}

type PreloadPreloadingAttemptSourcesUpdatedEvent struct {
	LoaderID                 NetworkLoaderID                  `json:"loaderId"`
	PreloadingAttemptSources []PreloadPreloadingAttemptSource `json:"preloadingAttemptSources"`
}

type ProfilerProfileNode struct {
	// Unique id of the node.
	ID int `json:"id"`
	// Function location.
	CallFrame RuntimeCallFrame `json:"callFrame"`
	// Number of samples where this node was on top of the call stack.
	HitCount *int `json:"hitCount,omitempty"`
	// Child node ids.
	Children []int `json:"children,omitempty"`
	// The reason of being not optimized. The function may be deoptimized or marked as don't
	DeoptReason *string `json:"deoptReason,omitempty"`
	// An array of source position ticks.
	PositionTicks []ProfilerPositionTickInfo `json:"positionTicks,omitempty"`
}

type ProfilerProfile struct {
	// The list of profile nodes. First item is the root node.
	Nodes []ProfilerProfileNode `json:"nodes"`
	// Profiling start timestamp in microseconds.
	StartTime float64 `json:"startTime"`
	// Profiling end timestamp in microseconds.
	EndTime float64 `json:"endTime"`
	// Ids of samples top nodes.
	Samples []int `json:"samples,omitempty"`
	// Time intervals between adjacent samples in microseconds. The first delta is relative to the
	TimeDeltas []int `json:"timeDeltas,omitempty"`
}

type ProfilerPositionTickInfo struct {
	// Source line number (1-based).
	Line int `json:"line"`
	// Number of samples attributed to the source line.
	Ticks int `json:"ticks"`
}

type ProfilerCoverageRange struct {
	// JavaScript script source offset for the range start.
	StartOffset int `json:"startOffset"`
	// JavaScript script source offset for the range end.
	EndOffset int `json:"endOffset"`
	// Collected execution count of the source range.
	Count int `json:"count"`
}

type ProfilerFunctionCoverage struct {
	// JavaScript function name.
	FunctionName string `json:"functionName"`
	// Source ranges inside the function with coverage data.
	Ranges []ProfilerCoverageRange `json:"ranges"`
	// Whether coverage data for this function has block granularity.
	IsBlockCoverage bool `json:"isBlockCoverage"`
}

type ProfilerScriptCoverage struct {
	// JavaScript script id.
	ScriptID RuntimeScriptID `json:"scriptId"`
	// JavaScript script name or url.
	URL string `json:"url"`
	// Functions contained in the script that has coverage data.
	Functions []ProfilerFunctionCoverage `json:"functions"`
}

type ProfilerDisableParams struct {
	SessionID string `json:"-"`
}

type ProfilerDisableResult struct {
}

type ProfilerEnableParams struct {
	SessionID string `json:"-"`
}

type ProfilerEnableResult struct {
}

type ProfilerGetBestEffortCoverageParams struct {
	SessionID string `json:"-"`
}

type ProfilerGetBestEffortCoverageResult struct {
	// Coverage data for the current isolate.
	Result []ProfilerScriptCoverage `json:"result"`
}

type ProfilerSetSamplingIntervalParams struct {
	SessionID string `json:"-"`
	// New sampling interval in microseconds.
	Interval int `json:"interval"`
}

type ProfilerSetSamplingIntervalResult struct {
}

type ProfilerStartParams struct {
	SessionID string `json:"-"`
}

type ProfilerStartResult struct {
}

type ProfilerStartPreciseCoverageParams struct {
	SessionID string `json:"-"`
	// Collect accurate call counts beyond simple 'covered' or 'not covered'.
	CallCount *bool `json:"callCount,omitempty"`
	// Collect block-based coverage.
	Detailed *bool `json:"detailed,omitempty"`
	// Allow the backend to send updates on its own initiative
	AllowTriggeredUpdates *bool `json:"allowTriggeredUpdates,omitempty"`
}

type ProfilerStartPreciseCoverageResult struct {
	// Monotonically increasing time (in seconds) when the coverage update was taken in the backend.
	Timestamp float64 `json:"timestamp"`
}

type ProfilerStopParams struct {
	SessionID string `json:"-"`
}

type ProfilerStopResult struct {
	// Recorded profile.
	Profile ProfilerProfile `json:"profile"`
}

type ProfilerStopPreciseCoverageParams struct {
	SessionID string `json:"-"`
}

type ProfilerStopPreciseCoverageResult struct {
}

type ProfilerTakePreciseCoverageParams struct {
	SessionID string `json:"-"`
}

type ProfilerTakePreciseCoverageResult struct {
	// Coverage data for the current isolate.
	Result []ProfilerScriptCoverage `json:"result"`
	// Monotonically increasing time (in seconds) when the coverage update was taken in the backend.
	Timestamp float64 `json:"timestamp"`
}

type ProfilerConsoleProfileFinishedEvent struct {
	ID string `json:"id"`
	// Location of console.profileEnd().
	Location DebuggerLocation `json:"location"`
	Profile  ProfilerProfile  `json:"profile"`
	// Profile title passed as an argument to console.profile().
	Title *string `json:"title,omitempty"`
}

type ProfilerConsoleProfileStartedEvent struct {
	ID string `json:"id"`
	// Location of console.profile().
	Location DebuggerLocation `json:"location"`
	// Profile title passed as an argument to console.profile().
	Title *string `json:"title,omitempty"`
}

type ProfilerPreciseCoverageDeltaUpdateEvent struct {
	// Monotonically increasing time (in seconds) when the coverage update was taken in the backend.
	Timestamp float64 `json:"timestamp"`
	// Identifier for distinguishing coverage events.
	Occasion string `json:"occasion"`
	// Coverage data for the current isolate.
	Result []ProfilerScriptCoverage `json:"result"`
}

type RuntimeScriptID string

type RuntimeSerializationOptions struct {
	Serialization string `json:"serialization"`
	// Deep serialization depth. Default is full depth. Respected only in `deep` serialization mode.
	MaxDepth *int `json:"maxDepth,omitempty"`
	// Embedder-specific parameters. For example if connected to V8 in Chrome these control DOM
	AdditionalParameters map[string]any `json:"additionalParameters,omitempty"`
}

type RuntimeDeepSerializedValue struct {
	Type     string  `json:"type"`
	Value    any     `json:"value,omitempty"`
	ObjectID *string `json:"objectId,omitempty"`
	// Set if value reference met more then once during serialization. In such
	WeakLocalObjectReference *int `json:"weakLocalObjectReference,omitempty"`
}

type RuntimeRemoteObjectID string

type RuntimeUnserializableValue string

type RuntimeRemoteObject struct {
	// Object type.
	Type string `json:"type"`
	// Object subtype hint. Specified for `object` type values only.
	Subtype *string `json:"subtype,omitempty"`
	// Object class (constructor) name. Specified for `object` type values only.
	ClassName *string `json:"className,omitempty"`
	// Remote object value in case of primitive values or JSON values (if it was requested).
	Value any `json:"value,omitempty"`
	// Primitive value which can not be JSON-stringified does not have `value`, but gets this
	UnserializableValue *RuntimeUnserializableValue `json:"unserializableValue,omitempty"`
	// String representation of the object.
	Description *string `json:"description,omitempty"`
	// Deep serialized value.
	DeepSerializedValue *RuntimeDeepSerializedValue `json:"deepSerializedValue,omitempty"`
	// Unique object identifier (for non-primitive values).
	ObjectID *RuntimeRemoteObjectID `json:"objectId,omitempty"`
	// Preview containing abbreviated property values. Specified for `object` type values only.
	Preview       *RuntimeObjectPreview `json:"preview,omitempty"`
	CustomPreview *RuntimeCustomPreview `json:"customPreview,omitempty"`
}

type RuntimeCustomPreview struct {
	// The JSON-stringified result of formatter.header(object, config) call.
	Header string `json:"header"`
	// If formatter returns true as a result of formatter.hasBody call then bodyGetterId will
	BodyGetterID *RuntimeRemoteObjectID `json:"bodyGetterId,omitempty"`
}

type RuntimeObjectPreview struct {
	// Object type.
	Type string `json:"type"`
	// Object subtype hint. Specified for `object` type values only.
	Subtype *string `json:"subtype,omitempty"`
	// String representation of the object.
	Description *string `json:"description,omitempty"`
	// True iff some of the properties or entries of the original object did not fit.
	Overflow bool `json:"overflow"`
	// List of the properties.
	Properties []RuntimePropertyPreview `json:"properties"`
	// List of the entries. Specified for `map` and `set` subtype values only.
	Entries []RuntimeEntryPreview `json:"entries,omitempty"`
}

type RuntimePropertyPreview struct {
	// Property name.
	Name string `json:"name"`
	// Object type. Accessor means that the property itself is an accessor property.
	Type string `json:"type"`
	// User-friendly property value string.
	Value *string `json:"value,omitempty"`
	// Nested value preview.
	ValuePreview *RuntimeObjectPreview `json:"valuePreview,omitempty"`
	// Object subtype hint. Specified for `object` type values only.
	Subtype *string `json:"subtype,omitempty"`
}

type RuntimeEntryPreview struct {
	// Preview of the key. Specified for map-like collection entries.
	Key *RuntimeObjectPreview `json:"key,omitempty"`
	// Preview of the value.
	Value RuntimeObjectPreview `json:"value"`
}

type RuntimePropertyDescriptor struct {
	// Property name or symbol description.
	Name string `json:"name"`
	// The value associated with the property.
	Value *RuntimeRemoteObject `json:"value,omitempty"`
	// True if the value associated with the property may be changed (data descriptors only).
	Writable *bool `json:"writable,omitempty"`
	// A function which serves as a getter for the property, or `undefined` if there is no getter
	Get *RuntimeRemoteObject `json:"get,omitempty"`
	// A function which serves as a setter for the property, or `undefined` if there is no setter
	Set *RuntimeRemoteObject `json:"set,omitempty"`
	// True if the type of this property descriptor may be changed and if the property may be
	Configurable bool `json:"configurable"`
	// True if this property shows up during enumeration of the properties on the corresponding
	Enumerable bool `json:"enumerable"`
	// True if the result was thrown during the evaluation.
	WasThrown *bool `json:"wasThrown,omitempty"`
	// True if the property is owned for the object.
	IsOwn *bool `json:"isOwn,omitempty"`
	// Property symbol object, if the property is of the `symbol` type.
	Symbol *RuntimeRemoteObject `json:"symbol,omitempty"`
}

type RuntimeInternalPropertyDescriptor struct {
	// Conventional property name.
	Name string `json:"name"`
	// The value associated with the property.
	Value *RuntimeRemoteObject `json:"value,omitempty"`
}

type RuntimePrivatePropertyDescriptor struct {
	// Private property name.
	Name string `json:"name"`
	// The value associated with the private property.
	Value *RuntimeRemoteObject `json:"value,omitempty"`
	// A function which serves as a getter for the private property,
	Get *RuntimeRemoteObject `json:"get,omitempty"`
	// A function which serves as a setter for the private property,
	Set *RuntimeRemoteObject `json:"set,omitempty"`
}

type RuntimeCallArgument struct {
	// Primitive value or serializable javascript object.
	Value any `json:"value,omitempty"`
	// Primitive value which can not be JSON-stringified.
	UnserializableValue *RuntimeUnserializableValue `json:"unserializableValue,omitempty"`
	// Remote object handle.
	ObjectID *RuntimeRemoteObjectID `json:"objectId,omitempty"`
}

type RuntimeExecutionContextID int

type RuntimeExecutionContextDescription struct {
	// Unique id of the execution context. It can be used to specify in which execution context
	ID RuntimeExecutionContextID `json:"id"`
	// Execution context origin.
	Origin string `json:"origin"`
	// Human readable name describing given context.
	Name string `json:"name"`
	// A system-unique execution context identifier. Unlike the id, this is unique across
	UniqueID string `json:"uniqueId"`
	// Embedder-specific auxiliary data likely matching {isDefault: boolean, type: 'default'|'isolated'|'worker', frameId: string}
	AuxData map[string]any `json:"auxData,omitempty"`
}

type RuntimeExceptionDetails struct {
	// Exception id.
	ExceptionID int `json:"exceptionId"`
	// Exception text, which should be used together with exception object when available.
	Text string `json:"text"`
	// Line number of the exception location (0-based).
	LineNumber int `json:"lineNumber"`
	// Column number of the exception location (0-based).
	ColumnNumber int `json:"columnNumber"`
	// Script ID of the exception location.
	ScriptID *RuntimeScriptID `json:"scriptId,omitempty"`
	// URL of the exception location, to be used when the script was not reported.
	URL *string `json:"url,omitempty"`
	// JavaScript stack trace if available.
	StackTrace *RuntimeStackTrace `json:"stackTrace,omitempty"`
	// Exception object if available.
	Exception *RuntimeRemoteObject `json:"exception,omitempty"`
	// Identifier of the context where exception happened.
	ExecutionContextID *RuntimeExecutionContextID `json:"executionContextId,omitempty"`
	// Dictionary with entries of meta data that the client associated
	ExceptionMetaData map[string]any `json:"exceptionMetaData,omitempty"`
}

type RuntimeTimestamp float64

type RuntimeTimeDelta float64

type RuntimeCallFrame struct {
	// JavaScript function name.
	FunctionName string `json:"functionName"`
	// JavaScript script id.
	ScriptID RuntimeScriptID `json:"scriptId"`
	// JavaScript script name or url.
	URL string `json:"url"`
	// JavaScript script line number (0-based).
	LineNumber int `json:"lineNumber"`
	// JavaScript script column number (0-based).
	ColumnNumber int `json:"columnNumber"`
}

type RuntimeStackTrace struct {
	// String label of this stack trace. For async traces this may be a name of the function that
	Description *string `json:"description,omitempty"`
	// JavaScript function name.
	CallFrames []RuntimeCallFrame `json:"callFrames"`
	// Asynchronous JavaScript stack trace that preceded this stack, if available.
	Parent *RuntimeStackTrace `json:"parent,omitempty"`
	// Asynchronous JavaScript stack trace that preceded this stack, if available.
	ParentID *RuntimeStackTraceID `json:"parentId,omitempty"`
}

type RuntimeUniqueDebuggerID string

type RuntimeStackTraceID struct {
	ID         string                   `json:"id"`
	DebuggerID *RuntimeUniqueDebuggerID `json:"debuggerId,omitempty"`
}

type RuntimeAwaitPromiseParams struct {
	SessionID string `json:"-"`
	// Identifier of the promise.
	PromiseObjectID RuntimeRemoteObjectID `json:"promiseObjectId"`
	// Whether the result is expected to be a JSON object that should be sent by value.
	ReturnByValue *bool `json:"returnByValue,omitempty"`
	// Whether preview should be generated for the result.
	GeneratePreview *bool `json:"generatePreview,omitempty"`
}

type RuntimeAwaitPromiseResult struct {
	// Promise result. Will contain rejected value if promise was rejected.
	Result RuntimeRemoteObject `json:"result"`
	// Exception details if stack strace is available.
	ExceptionDetails *RuntimeExceptionDetails `json:"exceptionDetails,omitempty"`
}

type RuntimeCallFunctionOnParams struct {
	SessionID string `json:"-"`
	// Declaration of the function to call.
	FunctionDeclaration string `json:"functionDeclaration"`
	// Identifier of the object to call function on. Either objectId or executionContextId should
	ObjectID *RuntimeRemoteObjectID `json:"objectId,omitempty"`
	// Call arguments. All call arguments must belong to the same JavaScript world as the target
	Arguments []RuntimeCallArgument `json:"arguments,omitempty"`
	// In silent mode exceptions thrown during evaluation are not reported and do not pause
	Silent *bool `json:"silent,omitempty"`
	// Whether the result is expected to be a JSON object which should be sent by value.
	ReturnByValue *bool `json:"returnByValue,omitempty"`
	// Whether preview should be generated for the result.
	GeneratePreview *bool `json:"generatePreview,omitempty"`
	// Whether execution should be treated as initiated by user in the UI.
	UserGesture *bool `json:"userGesture,omitempty"`
	// Whether execution should `await` for resulting value and return once awaited promise is
	AwaitPromise *bool `json:"awaitPromise,omitempty"`
	// Specifies execution context which global object will be used to call function on. Either
	ExecutionContextID *RuntimeExecutionContextID `json:"executionContextId,omitempty"`
	// Symbolic group name that can be used to release multiple objects. If objectGroup is not
	ObjectGroup *string `json:"objectGroup,omitempty"`
	// Whether to throw an exception if side effect cannot be ruled out during evaluation.
	ThrowOnSideEffect *bool `json:"throwOnSideEffect,omitempty"`
	// An alternative way to specify the execution context to call function on.
	UniqueContextID *string `json:"uniqueContextId,omitempty"`
	// Specifies the result serialization. If provided, overrides
	SerializationOptions *RuntimeSerializationOptions `json:"serializationOptions,omitempty"`
}

type RuntimeCallFunctionOnResult struct {
	// Call result.
	Result RuntimeRemoteObject `json:"result"`
	// Exception details.
	ExceptionDetails *RuntimeExceptionDetails `json:"exceptionDetails,omitempty"`
}

type RuntimeCompileScriptParams struct {
	SessionID string `json:"-"`
	// Expression to compile.
	Expression string `json:"expression"`
	// Source url to be set for the script.
	SourceURL string `json:"sourceURL"`
	// Specifies whether the compiled script should be persisted.
	PersistScript bool `json:"persistScript"`
	// Specifies in which execution context to perform script run. If the parameter is omitted the
	ExecutionContextID *RuntimeExecutionContextID `json:"executionContextId,omitempty"`
}

type RuntimeCompileScriptResult struct {
	// Id of the script.
	ScriptID *RuntimeScriptID `json:"scriptId,omitempty"`
	// Exception details.
	ExceptionDetails *RuntimeExceptionDetails `json:"exceptionDetails,omitempty"`
}

type RuntimeDisableParams struct {
	SessionID string `json:"-"`
}

type RuntimeDisableResult struct {
}

type RuntimeDiscardConsoleEntriesParams struct {
	SessionID string `json:"-"`
}

type RuntimeDiscardConsoleEntriesResult struct {
}

type RuntimeEnableParams struct {
	SessionID string `json:"-"`
}

type RuntimeEnableResult struct {
}

type RuntimeEvaluateParams struct {
	SessionID string `json:"-"`
	// Expression to evaluate.
	Expression string `json:"expression"`
	// Symbolic group name that can be used to release multiple objects.
	ObjectGroup *string `json:"objectGroup,omitempty"`
	// Determines whether Command Line API should be available during the evaluation.
	IncludeCommandLineAPI *bool `json:"includeCommandLineAPI,omitempty"`
	// In silent mode exceptions thrown during evaluation are not reported and do not pause
	Silent *bool `json:"silent,omitempty"`
	// Specifies in which execution context to perform evaluation. If the parameter is omitted the
	ContextID *RuntimeExecutionContextID `json:"contextId,omitempty"`
	// Whether the result is expected to be a JSON object that should be sent by value.
	ReturnByValue *bool `json:"returnByValue,omitempty"`
	// Whether preview should be generated for the result.
	GeneratePreview *bool `json:"generatePreview,omitempty"`
	// Whether execution should be treated as initiated by user in the UI.
	UserGesture *bool `json:"userGesture,omitempty"`
	// Whether execution should `await` for resulting value and return once awaited promise is
	AwaitPromise *bool `json:"awaitPromise,omitempty"`
	// Whether to throw an exception if side effect cannot be ruled out during evaluation.
	ThrowOnSideEffect *bool `json:"throwOnSideEffect,omitempty"`
	// Terminate execution after timing out (number of milliseconds).
	Timeout *RuntimeTimeDelta `json:"timeout,omitempty"`
	// Disable breakpoints during execution.
	DisableBreaks *bool `json:"disableBreaks,omitempty"`
	// Setting this flag to true enables `let` re-declaration and top-level `await`.
	ReplMode *bool `json:"replMode,omitempty"`
	// The Content Security Policy (CSP) for the target might block 'unsafe-eval'
	AllowUnsafeEvalBlockedByCSP *bool `json:"allowUnsafeEvalBlockedByCSP,omitempty"`
	// An alternative way to specify the execution context to evaluate in.
	UniqueContextID *string `json:"uniqueContextId,omitempty"`
	// Specifies the result serialization. If provided, overrides
	SerializationOptions *RuntimeSerializationOptions `json:"serializationOptions,omitempty"`
}

type RuntimeEvaluateResult struct {
	// Evaluation result.
	Result RuntimeRemoteObject `json:"result"`
	// Exception details.
	ExceptionDetails *RuntimeExceptionDetails `json:"exceptionDetails,omitempty"`
}

type RuntimeGetIsolateIDParams struct {
	SessionID string `json:"-"`
}

type RuntimeGetIsolateIDResult struct {
	// The isolate id.
	ID string `json:"id"`
}

type RuntimeGetHeapUsageParams struct {
	SessionID string `json:"-"`
}

type RuntimeGetHeapUsageResult struct {
	// Used JavaScript heap size in bytes.
	UsedSize float64 `json:"usedSize"`
	// Allocated JavaScript heap size in bytes.
	TotalSize float64 `json:"totalSize"`
	// Used size in bytes in the embedder's garbage-collected heap.
	EmbedderHeapUsedSize float64 `json:"embedderHeapUsedSize"`
	// Size in bytes of backing storage for array buffers and external strings.
	BackingStorageSize float64 `json:"backingStorageSize"`
}

type RuntimeGetPropertiesParams struct {
	SessionID string `json:"-"`
	// Identifier of the object to return properties for.
	ObjectID RuntimeRemoteObjectID `json:"objectId"`
	// If true, returns properties belonging only to the element itself, not to its prototype
	OwnProperties *bool `json:"ownProperties,omitempty"`
	// If true, returns accessor properties (with getter/setter) only; internal properties are not
	AccessorPropertiesOnly *bool `json:"accessorPropertiesOnly,omitempty"`
	// Whether preview should be generated for the results.
	GeneratePreview *bool `json:"generatePreview,omitempty"`
	// If true, returns non-indexed properties only.
	NonIndexedPropertiesOnly *bool `json:"nonIndexedPropertiesOnly,omitempty"`
}

type RuntimeGetPropertiesResult struct {
	// Object properties.
	Result []RuntimePropertyDescriptor `json:"result"`
	// Internal object properties (only of the element itself).
	InternalProperties []RuntimeInternalPropertyDescriptor `json:"internalProperties,omitempty"`
	// Object private properties.
	PrivateProperties []RuntimePrivatePropertyDescriptor `json:"privateProperties,omitempty"`
	// Exception details.
	ExceptionDetails *RuntimeExceptionDetails `json:"exceptionDetails,omitempty"`
}

type RuntimeGlobalLexicalScopeNamesParams struct {
	SessionID string `json:"-"`
	// Specifies in which execution context to lookup global scope variables.
	ExecutionContextID *RuntimeExecutionContextID `json:"executionContextId,omitempty"`
}

type RuntimeGlobalLexicalScopeNamesResult struct {
	Names []string `json:"names"`
}

type RuntimeQueryObjectsParams struct {
	SessionID string `json:"-"`
	// Identifier of the prototype to return objects for.
	PrototypeObjectID RuntimeRemoteObjectID `json:"prototypeObjectId"`
	// Symbolic group name that can be used to release the results.
	ObjectGroup *string `json:"objectGroup,omitempty"`
}

type RuntimeQueryObjectsResult struct {
	// Array with objects.
	Objects RuntimeRemoteObject `json:"objects"`
}

type RuntimeReleaseObjectParams struct {
	SessionID string `json:"-"`
	// Identifier of the object to release.
	ObjectID RuntimeRemoteObjectID `json:"objectId"`
}

type RuntimeReleaseObjectResult struct {
}

type RuntimeReleaseObjectGroupParams struct {
	SessionID string `json:"-"`
	// Symbolic object group name.
	ObjectGroup string `json:"objectGroup"`
}

type RuntimeReleaseObjectGroupResult struct {
}

type RuntimeRunIfWaitingForDebuggerParams struct {
	SessionID string `json:"-"`
}

type RuntimeRunIfWaitingForDebuggerResult struct {
}

type RuntimeRunScriptParams struct {
	SessionID string `json:"-"`
	// Id of the script to run.
	ScriptID RuntimeScriptID `json:"scriptId"`
	// Specifies in which execution context to perform script run. If the parameter is omitted the
	ExecutionContextID *RuntimeExecutionContextID `json:"executionContextId,omitempty"`
	// Symbolic group name that can be used to release multiple objects.
	ObjectGroup *string `json:"objectGroup,omitempty"`
	// In silent mode exceptions thrown during evaluation are not reported and do not pause
	Silent *bool `json:"silent,omitempty"`
	// Determines whether Command Line API should be available during the evaluation.
	IncludeCommandLineAPI *bool `json:"includeCommandLineAPI,omitempty"`
	// Whether the result is expected to be a JSON object which should be sent by value.
	ReturnByValue *bool `json:"returnByValue,omitempty"`
	// Whether preview should be generated for the result.
	GeneratePreview *bool `json:"generatePreview,omitempty"`
	// Whether execution should `await` for resulting value and return once awaited promise is
	AwaitPromise *bool `json:"awaitPromise,omitempty"`
}

type RuntimeRunScriptResult struct {
	// Run result.
	Result RuntimeRemoteObject `json:"result"`
	// Exception details.
	ExceptionDetails *RuntimeExceptionDetails `json:"exceptionDetails,omitempty"`
}

type RuntimeSetAsyncCallStackDepthParams struct {
	SessionID string `json:"-"`
	// Maximum depth of async call stacks. Setting to `0` will effectively disable collecting async
	MaxDepth int `json:"maxDepth"`
}

type RuntimeSetAsyncCallStackDepthResult struct {
}

type RuntimeSetCustomObjectFormatterEnabledParams struct {
	SessionID string `json:"-"`
	Enabled   bool   `json:"enabled"`
}

type RuntimeSetCustomObjectFormatterEnabledResult struct {
}

type RuntimeSetMaxCallStackSizeToCaptureParams struct {
	SessionID string `json:"-"`
	Size      int    `json:"size"`
}

type RuntimeSetMaxCallStackSizeToCaptureResult struct {
}

type RuntimeTerminateExecutionParams struct {
	SessionID string `json:"-"`
}

type RuntimeTerminateExecutionResult struct {
}

type RuntimeAddBindingParams struct {
	SessionID string `json:"-"`
	Name      string `json:"name"`
	// If specified, the binding would only be exposed to the specified
	ExecutionContextID *RuntimeExecutionContextID `json:"executionContextId,omitempty"`
	// If specified, the binding is exposed to the executionContext with
	ExecutionContextName *string `json:"executionContextName,omitempty"`
}

type RuntimeAddBindingResult struct {
}

type RuntimeRemoveBindingParams struct {
	SessionID string `json:"-"`
	Name      string `json:"name"`
}

type RuntimeRemoveBindingResult struct {
}

type RuntimeGetExceptionDetailsParams struct {
	SessionID string `json:"-"`
	// The error object for which to resolve the exception details.
	ErrorObjectID RuntimeRemoteObjectID `json:"errorObjectId"`
}

type RuntimeGetExceptionDetailsResult struct {
	ExceptionDetails *RuntimeExceptionDetails `json:"exceptionDetails,omitempty"`
}

type RuntimeBindingCalledEvent struct {
	Name    string `json:"name"`
	Payload string `json:"payload"`
	// Identifier of the context where the call was made.
	ExecutionContextID RuntimeExecutionContextID `json:"executionContextId"`
}

type RuntimeConsoleAPICalledEvent struct {
	// Type of the call.
	Type string `json:"type"`
	// Call arguments.
	Args []RuntimeRemoteObject `json:"args"`
	// Identifier of the context where the call was made.
	ExecutionContextID RuntimeExecutionContextID `json:"executionContextId"`
	// Call timestamp.
	Timestamp RuntimeTimestamp `json:"timestamp"`
	// Stack trace captured when the call was made. The async stack chain is automatically reported for
	StackTrace *RuntimeStackTrace `json:"stackTrace,omitempty"`
	// Console context descriptor for calls on non-default console context (not console.*):
	Context *string `json:"context,omitempty"`
}

type RuntimeExceptionRevokedEvent struct {
	// Reason describing why exception was revoked.
	Reason string `json:"reason"`
	// The id of revoked exception, as reported in `exceptionThrown`.
	ExceptionID int `json:"exceptionId"`
}

type RuntimeExceptionThrownEvent struct {
	// Timestamp of the exception.
	Timestamp        RuntimeTimestamp        `json:"timestamp"`
	ExceptionDetails RuntimeExceptionDetails `json:"exceptionDetails"`
}

type RuntimeExecutionContextCreatedEvent struct {
	// A newly created execution context.
	Context RuntimeExecutionContextDescription `json:"context"`
}

type RuntimeExecutionContextDestroyedEvent struct {
	// Id of the destroyed context
	ExecutionContextID RuntimeExecutionContextID `json:"executionContextId"`
	// Unique Id of the destroyed context
	ExecutionContextUniqueID string `json:"executionContextUniqueId"`
}

type RuntimeExecutionContextsClearedEvent struct {
}

type RuntimeInspectRequestedEvent struct {
	Object RuntimeRemoteObject `json:"object"`
	Hints  map[string]any      `json:"hints"`
	// Identifier of the context where the call was made.
	ExecutionContextID *RuntimeExecutionContextID `json:"executionContextId,omitempty"`
}

type SchemaDomainType struct {
	// Domain name.
	Name string `json:"name"`
	// Domain version.
	Version string `json:"version"`
}

type SchemaGetDomainsParams struct {
	SessionID string `json:"-"`
}

type SchemaGetDomainsResult struct {
	// List of supported domains.
	Domains []SchemaDomainType `json:"domains"`
}

type SecurityCertificateID int

type SecurityMixedContentType string

type SecuritySecurityState string

type SecurityCertificateSecurityState struct {
	// Protocol name (e.g. "TLS 1.2" or "QUIC").
	Protocol string `json:"protocol"`
	// Key Exchange used by the connection, or the empty string if not applicable.
	KeyExchange string `json:"keyExchange"`
	// (EC)DH group used by the connection, if applicable.
	KeyExchangeGroup *string `json:"keyExchangeGroup,omitempty"`
	// Cipher name.
	Cipher string `json:"cipher"`
	// TLS MAC. Note that AEAD ciphers do not have separate MACs.
	Mac *string `json:"mac,omitempty"`
	// Page certificate.
	Certificate []string `json:"certificate"`
	// Certificate subject name.
	SubjectName string `json:"subjectName"`
	// Name of the issuing CA.
	Issuer string `json:"issuer"`
	// Certificate valid from date.
	ValidFrom NetworkTimeSinceEpoch `json:"validFrom"`
	// Certificate valid to (expiration) date
	ValidTo NetworkTimeSinceEpoch `json:"validTo"`
	// The highest priority network error code, if the certificate has an error.
	CertificateNetworkError *string `json:"certificateNetworkError,omitempty"`
	// True if the certificate uses a weak signature algorithm.
	CertificateHasWeakSignature bool `json:"certificateHasWeakSignature"`
	// True if the certificate has a SHA1 signature in the chain.
	CertificateHasSha1Signature bool `json:"certificateHasSha1Signature"`
	// True if modern SSL
	ModernSSL bool `json:"modernSSL"`
	// True if the connection is using an obsolete SSL protocol.
	ObsoleteSslProtocol bool `json:"obsoleteSslProtocol"`
	// True if the connection is using an obsolete SSL key exchange.
	ObsoleteSslKeyExchange bool `json:"obsoleteSslKeyExchange"`
	// True if the connection is using an obsolete SSL cipher.
	ObsoleteSslCipher bool `json:"obsoleteSslCipher"`
	// True if the connection is using an obsolete SSL signature.
	ObsoleteSslSignature bool `json:"obsoleteSslSignature"`
}

type SecuritySafetyTipStatus string

type SecuritySafetyTipInfo struct {
	// Describes whether the page triggers any safety tips or reputation warnings. Default is unknown.
	SafetyTipStatus SecuritySafetyTipStatus `json:"safetyTipStatus"`
	// The URL the safety tip suggested ("Did you mean?"). Only filled in for lookalike matches.
	SafeURL *string `json:"safeUrl,omitempty"`
}

type SecurityVisibleSecurityState struct {
	// The security level of the page.
	SecurityState SecuritySecurityState `json:"securityState"`
	// Security state details about the page certificate.
	CertificateSecurityState *SecurityCertificateSecurityState `json:"certificateSecurityState,omitempty"`
	// The type of Safety Tip triggered on the page. Note that this field will be set even if the Safety Tip UI was not actually shown.
	SafetyTipInfo *SecuritySafetyTipInfo `json:"safetyTipInfo,omitempty"`
	// Array of security state issues ids.
	SecurityStateIssueIds []string `json:"securityStateIssueIds"`
}

type SecuritySecurityStateExplanation struct {
	// Security state representing the severity of the factor being explained.
	SecurityState SecuritySecurityState `json:"securityState"`
	// Title describing the type of factor.
	Title string `json:"title"`
	// Short phrase describing the type of factor.
	Summary string `json:"summary"`
	// Full text explanation of the factor.
	Description string `json:"description"`
	// The type of mixed content described by the explanation.
	MixedContentType SecurityMixedContentType `json:"mixedContentType"`
	// Page certificate.
	Certificate []string `json:"certificate"`
	// Recommendations to fix any issues.
	Recommendations []string `json:"recommendations,omitempty"`
}

type SecurityInsecureContentStatus struct {
	// Always false.
	RanMixedContent bool `json:"ranMixedContent"`
	// Always false.
	DisplayedMixedContent bool `json:"displayedMixedContent"`
	// Always false.
	ContainedMixedForm bool `json:"containedMixedForm"`
	// Always false.
	RanContentWithCertErrors bool `json:"ranContentWithCertErrors"`
	// Always false.
	DisplayedContentWithCertErrors bool `json:"displayedContentWithCertErrors"`
	// Always set to unknown.
	RanInsecureContentStyle SecuritySecurityState `json:"ranInsecureContentStyle"`
	// Always set to unknown.
	DisplayedInsecureContentStyle SecuritySecurityState `json:"displayedInsecureContentStyle"`
}

type SecurityCertificateErrorAction string

type SecurityDisableParams struct {
	SessionID string `json:"-"`
}

type SecurityDisableResult struct {
}

type SecurityEnableParams struct {
	SessionID string `json:"-"`
}

type SecurityEnableResult struct {
}

type SecuritySetIgnoreCertificateErrorsParams struct {
	SessionID string `json:"-"`
	// If true, all certificate errors will be ignored.
	Ignore bool `json:"ignore"`
}

type SecuritySetIgnoreCertificateErrorsResult struct {
}

type SecurityHandleCertificateErrorParams struct {
	SessionID string `json:"-"`
	// The ID of the event.
	EventID int `json:"eventId"`
	// The action to take on the certificate error.
	Action SecurityCertificateErrorAction `json:"action"`
}

type SecurityHandleCertificateErrorResult struct {
}

type SecuritySetOverrideCertificateErrorsParams struct {
	SessionID string `json:"-"`
	// If true, certificate errors will be overridden.
	Override bool `json:"override"`
}

type SecuritySetOverrideCertificateErrorsResult struct {
}

type SecurityCertificateErrorEvent struct {
	// The ID of the event.
	EventID int `json:"eventId"`
	// The type of the error.
	ErrorType string `json:"errorType"`
	// The url that was requested.
	RequestURL string `json:"requestURL"`
}

type SecurityVisibleSecurityStateChangedEvent struct {
	// Security state information about the page.
	VisibleSecurityState SecurityVisibleSecurityState `json:"visibleSecurityState"`
}

type SecuritySecurityStateChangedEvent struct {
	// Security state.
	SecurityState SecuritySecurityState `json:"securityState"`
	// True if the page was loaded over cryptographic transport such as HTTPS.
	SchemeIsCryptographic bool `json:"schemeIsCryptographic"`
	// Previously a list of explanations for the security state. Now always
	Explanations []SecuritySecurityStateExplanation `json:"explanations"`
	// Information about insecure content on the page.
	InsecureContentStatus SecurityInsecureContentStatus `json:"insecureContentStatus"`
	// Overrides user-visible description of the state. Always omitted.
	Summary *string `json:"summary,omitempty"`
}

type ServiceWorkerRegistrationID string

type ServiceWorkerServiceWorkerRegistration struct {
	RegistrationID ServiceWorkerRegistrationID `json:"registrationId"`
	ScopeURL       string                      `json:"scopeURL"`
	IsDeleted      bool                        `json:"isDeleted"`
}

type ServiceWorkerServiceWorkerVersionRunningStatus string

type ServiceWorkerServiceWorkerVersionStatus string

type ServiceWorkerServiceWorkerVersion struct {
	VersionID      string                                         `json:"versionId"`
	RegistrationID ServiceWorkerRegistrationID                    `json:"registrationId"`
	ScriptURL      string                                         `json:"scriptURL"`
	RunningStatus  ServiceWorkerServiceWorkerVersionRunningStatus `json:"runningStatus"`
	Status         ServiceWorkerServiceWorkerVersionStatus        `json:"status"`
	// The Last-Modified header value of the main script.
	ScriptLastModified *float64 `json:"scriptLastModified,omitempty"`
	// The time at which the response headers of the main script were received from the server.
	ScriptResponseTime *float64         `json:"scriptResponseTime,omitempty"`
	ControlledClients  []TargetTargetID `json:"controlledClients,omitempty"`
	TargetID           *TargetTargetID  `json:"targetId,omitempty"`
	RouterRules        *string          `json:"routerRules,omitempty"`
}

type ServiceWorkerServiceWorkerErrorMessage struct {
	ErrorMessage   string                      `json:"errorMessage"`
	RegistrationID ServiceWorkerRegistrationID `json:"registrationId"`
	VersionID      string                      `json:"versionId"`
	SourceURL      string                      `json:"sourceURL"`
	LineNumber     int                         `json:"lineNumber"`
	ColumnNumber   int                         `json:"columnNumber"`
}

type ServiceWorkerDeliverPushMessageParams struct {
	SessionID      string                      `json:"-"`
	Origin         string                      `json:"origin"`
	RegistrationID ServiceWorkerRegistrationID `json:"registrationId"`
	Data           string                      `json:"data"`
}

type ServiceWorkerDeliverPushMessageResult struct {
}

type ServiceWorkerDisableParams struct {
	SessionID string `json:"-"`
}

type ServiceWorkerDisableResult struct {
}

type ServiceWorkerDispatchSyncEventParams struct {
	SessionID      string                      `json:"-"`
	Origin         string                      `json:"origin"`
	RegistrationID ServiceWorkerRegistrationID `json:"registrationId"`
	Tag            string                      `json:"tag"`
	LastChance     bool                        `json:"lastChance"`
}

type ServiceWorkerDispatchSyncEventResult struct {
}

type ServiceWorkerDispatchPeriodicSyncEventParams struct {
	SessionID      string                      `json:"-"`
	Origin         string                      `json:"origin"`
	RegistrationID ServiceWorkerRegistrationID `json:"registrationId"`
	Tag            string                      `json:"tag"`
}

type ServiceWorkerDispatchPeriodicSyncEventResult struct {
}

type ServiceWorkerEnableParams struct {
	SessionID string `json:"-"`
}

type ServiceWorkerEnableResult struct {
}

type ServiceWorkerSetForceUpdateOnPageLoadParams struct {
	SessionID             string `json:"-"`
	ForceUpdateOnPageLoad bool   `json:"forceUpdateOnPageLoad"`
}

type ServiceWorkerSetForceUpdateOnPageLoadResult struct {
}

type ServiceWorkerSkipWaitingParams struct {
	SessionID string `json:"-"`
	ScopeURL  string `json:"scopeURL"`
}

type ServiceWorkerSkipWaitingResult struct {
}

type ServiceWorkerStartWorkerParams struct {
	SessionID string `json:"-"`
	ScopeURL  string `json:"scopeURL"`
}

type ServiceWorkerStartWorkerResult struct {
}

type ServiceWorkerStopAllWorkersParams struct {
	SessionID string `json:"-"`
}

type ServiceWorkerStopAllWorkersResult struct {
}

type ServiceWorkerStopWorkerParams struct {
	SessionID string `json:"-"`
	VersionID string `json:"versionId"`
}

type ServiceWorkerStopWorkerResult struct {
}

type ServiceWorkerUnregisterParams struct {
	SessionID string `json:"-"`
	ScopeURL  string `json:"scopeURL"`
}

type ServiceWorkerUnregisterResult struct {
}

type ServiceWorkerUpdateRegistrationParams struct {
	SessionID string `json:"-"`
	ScopeURL  string `json:"scopeURL"`
}

type ServiceWorkerUpdateRegistrationResult struct {
}

type ServiceWorkerWorkerErrorReportedEvent struct {
	ErrorMessage ServiceWorkerServiceWorkerErrorMessage `json:"errorMessage"`
}

type ServiceWorkerWorkerRegistrationUpdatedEvent struct {
	Registrations []ServiceWorkerServiceWorkerRegistration `json:"registrations"`
}

type ServiceWorkerWorkerVersionUpdatedEvent struct {
	Versions []ServiceWorkerServiceWorkerVersion `json:"versions"`
}

type SmartCardEmulationResultCode string

type SmartCardEmulationShareMode string

type SmartCardEmulationDisposition string

type SmartCardEmulationConnectionState string

type SmartCardEmulationReaderStateFlags struct {
	Unaware     *bool `json:"unaware,omitempty"`
	Ignore      *bool `json:"ignore,omitempty"`
	Changed     *bool `json:"changed,omitempty"`
	Unknown     *bool `json:"unknown,omitempty"`
	Unavailable *bool `json:"unavailable,omitempty"`
	Empty       *bool `json:"empty,omitempty"`
	Present     *bool `json:"present,omitempty"`
	Exclusive   *bool `json:"exclusive,omitempty"`
	Inuse       *bool `json:"inuse,omitempty"`
	Mute        *bool `json:"mute,omitempty"`
	Unpowered   *bool `json:"unpowered,omitempty"`
}

type SmartCardEmulationProtocolSet struct {
	T0  *bool `json:"t0,omitempty"`
	T1  *bool `json:"t1,omitempty"`
	Raw *bool `json:"raw,omitempty"`
}

type SmartCardEmulationProtocol string

type SmartCardEmulationReaderStateIn struct {
	Reader                string                             `json:"reader"`
	CurrentState          SmartCardEmulationReaderStateFlags `json:"currentState"`
	CurrentInsertionCount int                                `json:"currentInsertionCount"`
}

type SmartCardEmulationReaderStateOut struct {
	Reader     string                             `json:"reader"`
	EventState SmartCardEmulationReaderStateFlags `json:"eventState"`
	EventCount int                                `json:"eventCount"`
	Atr        string                             `json:"atr"`
}

type SmartCardEmulationEnableParams struct {
	SessionID string `json:"-"`
}

type SmartCardEmulationEnableResult struct {
}

type SmartCardEmulationDisableParams struct {
	SessionID string `json:"-"`
}

type SmartCardEmulationDisableResult struct {
}

type SmartCardEmulationReportEstablishContextResultParams struct {
	SessionID string `json:"-"`
	RequestID string `json:"requestId"`
	ContextID int    `json:"contextId"`
}

type SmartCardEmulationReportEstablishContextResultResult struct {
}

type SmartCardEmulationReportReleaseContextResultParams struct {
	SessionID string `json:"-"`
	RequestID string `json:"requestId"`
}

type SmartCardEmulationReportReleaseContextResultResult struct {
}

type SmartCardEmulationReportListReadersResultParams struct {
	SessionID string   `json:"-"`
	RequestID string   `json:"requestId"`
	Readers   []string `json:"readers"`
}

type SmartCardEmulationReportListReadersResultResult struct {
}

type SmartCardEmulationReportGetStatusChangeResultParams struct {
	SessionID    string                             `json:"-"`
	RequestID    string                             `json:"requestId"`
	ReaderStates []SmartCardEmulationReaderStateOut `json:"readerStates"`
}

type SmartCardEmulationReportGetStatusChangeResultResult struct {
}

type SmartCardEmulationReportBeginTransactionResultParams struct {
	SessionID string `json:"-"`
	RequestID string `json:"requestId"`
	Handle    int    `json:"handle"`
}

type SmartCardEmulationReportBeginTransactionResultResult struct {
}

type SmartCardEmulationReportPlainResultParams struct {
	SessionID string `json:"-"`
	RequestID string `json:"requestId"`
}

type SmartCardEmulationReportPlainResultResult struct {
}

type SmartCardEmulationReportConnectResultParams struct {
	SessionID      string                      `json:"-"`
	RequestID      string                      `json:"requestId"`
	Handle         int                         `json:"handle"`
	ActiveProtocol *SmartCardEmulationProtocol `json:"activeProtocol,omitempty"`
}

type SmartCardEmulationReportConnectResultResult struct {
}

type SmartCardEmulationReportDataResultParams struct {
	SessionID string `json:"-"`
	RequestID string `json:"requestId"`
	Data      string `json:"data"`
}

type SmartCardEmulationReportDataResultResult struct {
}

type SmartCardEmulationReportStatusResultParams struct {
	SessionID  string                            `json:"-"`
	RequestID  string                            `json:"requestId"`
	ReaderName string                            `json:"readerName"`
	State      SmartCardEmulationConnectionState `json:"state"`
	Atr        string                            `json:"atr"`
	Protocol   *SmartCardEmulationProtocol       `json:"protocol,omitempty"`
}

type SmartCardEmulationReportStatusResultResult struct {
}

type SmartCardEmulationReportErrorParams struct {
	SessionID  string                       `json:"-"`
	RequestID  string                       `json:"requestId"`
	ResultCode SmartCardEmulationResultCode `json:"resultCode"`
}

type SmartCardEmulationReportErrorResult struct {
}

type SmartCardEmulationEstablishContextRequestedEvent struct {
	RequestID string `json:"requestId"`
}

type SmartCardEmulationReleaseContextRequestedEvent struct {
	RequestID string `json:"requestId"`
	ContextID int    `json:"contextId"`
}

type SmartCardEmulationListReadersRequestedEvent struct {
	RequestID string `json:"requestId"`
	ContextID int    `json:"contextId"`
}

type SmartCardEmulationGetStatusChangeRequestedEvent struct {
	RequestID    string                            `json:"requestId"`
	ContextID    int                               `json:"contextId"`
	ReaderStates []SmartCardEmulationReaderStateIn `json:"readerStates"`
	// in milliseconds, if absent, it means "infinite"
	Timeout *int `json:"timeout,omitempty"`
}

type SmartCardEmulationCancelRequestedEvent struct {
	RequestID string `json:"requestId"`
	ContextID int    `json:"contextId"`
}

type SmartCardEmulationConnectRequestedEvent struct {
	RequestID          string                        `json:"requestId"`
	ContextID          int                           `json:"contextId"`
	Reader             string                        `json:"reader"`
	ShareMode          SmartCardEmulationShareMode   `json:"shareMode"`
	PreferredProtocols SmartCardEmulationProtocolSet `json:"preferredProtocols"`
}

type SmartCardEmulationDisconnectRequestedEvent struct {
	RequestID   string                        `json:"requestId"`
	Handle      int                           `json:"handle"`
	Disposition SmartCardEmulationDisposition `json:"disposition"`
}

type SmartCardEmulationTransmitRequestedEvent struct {
	RequestID string                      `json:"requestId"`
	Handle    int                         `json:"handle"`
	Data      string                      `json:"data"`
	Protocol  *SmartCardEmulationProtocol `json:"protocol,omitempty"`
}

type SmartCardEmulationControlRequestedEvent struct {
	RequestID   string `json:"requestId"`
	Handle      int    `json:"handle"`
	ControlCode int    `json:"controlCode"`
	Data        string `json:"data"`
}

type SmartCardEmulationGetAttribRequestedEvent struct {
	RequestID string `json:"requestId"`
	Handle    int    `json:"handle"`
	AttribID  int    `json:"attribId"`
}

type SmartCardEmulationSetAttribRequestedEvent struct {
	RequestID string `json:"requestId"`
	Handle    int    `json:"handle"`
	AttribID  int    `json:"attribId"`
	Data      string `json:"data"`
}

type SmartCardEmulationStatusRequestedEvent struct {
	RequestID string `json:"requestId"`
	Handle    int    `json:"handle"`
}

type SmartCardEmulationBeginTransactionRequestedEvent struct {
	RequestID string `json:"requestId"`
	Handle    int    `json:"handle"`
}

type SmartCardEmulationEndTransactionRequestedEvent struct {
	RequestID   string                        `json:"requestId"`
	Handle      int                           `json:"handle"`
	Disposition SmartCardEmulationDisposition `json:"disposition"`
}

type StorageSerializedStorageKey string

type StorageStorageType string

type StorageUsageForType struct {
	// Name of storage type.
	StorageType StorageStorageType `json:"storageType"`
	// Storage usage (bytes).
	Usage float64 `json:"usage"`
}

type StorageTrustTokens struct {
	IssuerOrigin string  `json:"issuerOrigin"`
	Count        float64 `json:"count"`
}

type StorageInterestGroupAuctionID string

type StorageInterestGroupAccessType string

type StorageInterestGroupAuctionEventType string

type StorageInterestGroupAuctionFetchType string

type StorageSharedStorageAccessScope string

type StorageSharedStorageAccessMethod string

type StorageSharedStorageEntry struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type StorageSharedStorageMetadata struct {
	// Time when the origin's shared storage was last created.
	CreationTime NetworkTimeSinceEpoch `json:"creationTime"`
	// Number of key-value pairs stored in origin's shared storage.
	Length int `json:"length"`
	// Current amount of bits of entropy remaining in the navigation budget.
	RemainingBudget float64 `json:"remainingBudget"`
	// Total number of bytes stored as key-value pairs in origin's shared
	BytesUsed int `json:"bytesUsed"`
}

type StorageSharedStoragePrivateAggregationConfig struct {
	// The chosen aggregation service deployment.
	AggregationCoordinatorOrigin *string `json:"aggregationCoordinatorOrigin,omitempty"`
	// The context ID provided.
	ContextID *string `json:"contextId,omitempty"`
	// Configures the maximum size allowed for filtering IDs.
	FilteringIDMaxBytes int `json:"filteringIdMaxBytes"`
	// The limit on the number of contributions in the final report.
	MaxContributions *int `json:"maxContributions,omitempty"`
}

type StorageSharedStorageReportingMetadata struct {
	EventType    string `json:"eventType"`
	ReportingURL string `json:"reportingUrl"`
}

type StorageSharedStorageURLWithMetadata struct {
	// Spec of candidate URL.
	URL string `json:"url"`
	// Any associated reporting metadata.
	ReportingMetadata []StorageSharedStorageReportingMetadata `json:"reportingMetadata"`
}

type StorageSharedStorageAccessParams struct {
	// Spec of the module script URL.
	ScriptSourceURL *string `json:"scriptSourceUrl,omitempty"`
	// String denoting "context-origin", "script-origin", or a custom
	DataOrigin *string `json:"dataOrigin,omitempty"`
	// Name of the registered operation to be run.
	OperationName *string `json:"operationName,omitempty"`
	// ID of the operation call.
	OperationID *string `json:"operationId,omitempty"`
	// Whether or not to keep the worket alive for future run or selectURL
	KeepAlive *bool `json:"keepAlive,omitempty"`
	// Configures the private aggregation options.
	PrivateAggregationConfig *StorageSharedStoragePrivateAggregationConfig `json:"privateAggregationConfig,omitempty"`
	// The operation's serialized data in bytes (converted to a string).
	SerializedData *string `json:"serializedData,omitempty"`
	// Array of candidate URLs' specs, along with any associated metadata.
	UrlsWithMetadata []StorageSharedStorageURLWithMetadata `json:"urlsWithMetadata,omitempty"`
	// Spec of the URN:UUID generated for a selectURL call.
	UrnUUID *string `json:"urnUuid,omitempty"`
	// Key for a specific entry in an origin's shared storage.
	Key *string `json:"key,omitempty"`
	// Value for a specific entry in an origin's shared storage.
	Value *string `json:"value,omitempty"`
	// Whether or not to set an entry for a key if that key is already present.
	IgnoreIfPresent *bool `json:"ignoreIfPresent,omitempty"`
	// A number denoting the (0-based) order of the worklet's
	WorkletOrdinal *int `json:"workletOrdinal,omitempty"`
	// Hex representation of the DevTools token used as the TargetID for the
	WorkletTargetID *TargetTargetID `json:"workletTargetId,omitempty"`
	// Name of the lock to be acquired, if present.
	WithLock *string `json:"withLock,omitempty"`
	// If the method has been called as part of a batchUpdate, then this
	BatchUpdateID *string `json:"batchUpdateId,omitempty"`
	// Number of modifier methods sent in batch.
	BatchSize *int `json:"batchSize,omitempty"`
}

type StorageStorageBucketsDurability string

type StorageStorageBucket struct {
	StorageKey StorageSerializedStorageKey `json:"storageKey"`
	// If not specified, it is the default bucket of the storageKey.
	Name *string `json:"name,omitempty"`
}

type StorageStorageBucketInfo struct {
	Bucket     StorageStorageBucket  `json:"bucket"`
	ID         string                `json:"id"`
	Expiration NetworkTimeSinceEpoch `json:"expiration"`
	// Storage quota (bytes).
	Quota      float64                         `json:"quota"`
	Persistent bool                            `json:"persistent"`
	Durability StorageStorageBucketsDurability `json:"durability"`
}

type StorageAttributionReportingSourceType string

type StorageUnsignedInt64AsBase10 string

type StorageUnsignedInt128AsBase16 string

type StorageSignedInt64AsBase10 string

type StorageAttributionReportingFilterDataEntry struct {
	Key    string   `json:"key"`
	Values []string `json:"values"`
}

type StorageAttributionReportingFilterConfig struct {
	FilterValues []StorageAttributionReportingFilterDataEntry `json:"filterValues"`
	// duration in seconds
	LookbackWindow *int `json:"lookbackWindow,omitempty"`
}

type StorageAttributionReportingFilterPair struct {
	Filters    []StorageAttributionReportingFilterConfig `json:"filters"`
	NotFilters []StorageAttributionReportingFilterConfig `json:"notFilters"`
}

type StorageAttributionReportingAggregationKeysEntry struct {
	Key   string                        `json:"key"`
	Value StorageUnsignedInt128AsBase16 `json:"value"`
}

type StorageAttributionReportingEventReportWindows struct {
	// duration in seconds
	Start int `json:"start"`
	// duration in seconds
	Ends []int `json:"ends"`
}

type StorageAttributionReportingTriggerDataMatching string

type StorageAttributionReportingAggregatableDebugReportingData struct {
	KeyPiece StorageUnsignedInt128AsBase16 `json:"keyPiece"`
	// number instead of integer because not all uint32 can be represented by
	Value float64  `json:"value"`
	Types []string `json:"types"`
}

type StorageAttributionReportingAggregatableDebugReportingConfig struct {
	// number instead of integer because not all uint32 can be represented by
	Budget                       *float64                                                    `json:"budget,omitempty"`
	KeyPiece                     StorageUnsignedInt128AsBase16                               `json:"keyPiece"`
	DebugData                    []StorageAttributionReportingAggregatableDebugReportingData `json:"debugData"`
	AggregationCoordinatorOrigin *string                                                     `json:"aggregationCoordinatorOrigin,omitempty"`
}

type StorageAttributionScopesData struct {
	Values []string `json:"values"`
	// number instead of integer because not all uint32 can be represented by
	Limit          float64 `json:"limit"`
	MaxEventStates float64 `json:"maxEventStates"`
}

type StorageAttributionReportingNamedBudgetDef struct {
	Name   string `json:"name"`
	Budget int    `json:"budget"`
}

type StorageAttributionReportingSourceRegistration struct {
	Time NetworkTimeSinceEpoch `json:"time"`
	// duration in seconds
	Expiry int `json:"expiry"`
	// number instead of integer because not all uint32 can be represented by
	TriggerData        []float64                                     `json:"triggerData"`
	EventReportWindows StorageAttributionReportingEventReportWindows `json:"eventReportWindows"`
	// duration in seconds
	AggregatableReportWindow         int                                                         `json:"aggregatableReportWindow"`
	Type                             StorageAttributionReportingSourceType                       `json:"type"`
	SourceOrigin                     string                                                      `json:"sourceOrigin"`
	ReportingOrigin                  string                                                      `json:"reportingOrigin"`
	DestinationSites                 []string                                                    `json:"destinationSites"`
	EventID                          StorageUnsignedInt64AsBase10                                `json:"eventId"`
	Priority                         StorageSignedInt64AsBase10                                  `json:"priority"`
	FilterData                       []StorageAttributionReportingFilterDataEntry                `json:"filterData"`
	AggregationKeys                  []StorageAttributionReportingAggregationKeysEntry           `json:"aggregationKeys"`
	DebugKey                         *StorageUnsignedInt64AsBase10                               `json:"debugKey,omitempty"`
	TriggerDataMatching              StorageAttributionReportingTriggerDataMatching              `json:"triggerDataMatching"`
	DestinationLimitPriority         StorageSignedInt64AsBase10                                  `json:"destinationLimitPriority"`
	AggregatableDebugReportingConfig StorageAttributionReportingAggregatableDebugReportingConfig `json:"aggregatableDebugReportingConfig"`
	ScopesData                       *StorageAttributionScopesData                               `json:"scopesData,omitempty"`
	MaxEventLevelReports             int                                                         `json:"maxEventLevelReports"`
	NamedBudgets                     []StorageAttributionReportingNamedBudgetDef                 `json:"namedBudgets"`
	DebugReporting                   bool                                                        `json:"debugReporting"`
	EventLevelEpsilon                float64                                                     `json:"eventLevelEpsilon"`
}

type StorageAttributionReportingSourceRegistrationResult string

type StorageAttributionReportingSourceRegistrationTimeConfig string

type StorageAttributionReportingAggregatableValueDictEntry struct {
	Key string `json:"key"`
	// number instead of integer because not all uint32 can be represented by
	Value       float64                      `json:"value"`
	FilteringID StorageUnsignedInt64AsBase10 `json:"filteringId"`
}

type StorageAttributionReportingAggregatableValueEntry struct {
	Values  []StorageAttributionReportingAggregatableValueDictEntry `json:"values"`
	Filters StorageAttributionReportingFilterPair                   `json:"filters"`
}

type StorageAttributionReportingEventTriggerData struct {
	Data     StorageUnsignedInt64AsBase10          `json:"data"`
	Priority StorageSignedInt64AsBase10            `json:"priority"`
	DedupKey *StorageUnsignedInt64AsBase10         `json:"dedupKey,omitempty"`
	Filters  StorageAttributionReportingFilterPair `json:"filters"`
}

type StorageAttributionReportingAggregatableTriggerData struct {
	KeyPiece   StorageUnsignedInt128AsBase16         `json:"keyPiece"`
	SourceKeys []string                              `json:"sourceKeys"`
	Filters    StorageAttributionReportingFilterPair `json:"filters"`
}

type StorageAttributionReportingAggregatableDedupKey struct {
	DedupKey *StorageUnsignedInt64AsBase10         `json:"dedupKey,omitempty"`
	Filters  StorageAttributionReportingFilterPair `json:"filters"`
}

type StorageAttributionReportingNamedBudgetCandidate struct {
	Name    *string                               `json:"name,omitempty"`
	Filters StorageAttributionReportingFilterPair `json:"filters"`
}

type StorageAttributionReportingTriggerRegistration struct {
	Filters                          StorageAttributionReportingFilterPair                       `json:"filters"`
	DebugKey                         *StorageUnsignedInt64AsBase10                               `json:"debugKey,omitempty"`
	AggregatableDedupKeys            []StorageAttributionReportingAggregatableDedupKey           `json:"aggregatableDedupKeys"`
	EventTriggerData                 []StorageAttributionReportingEventTriggerData               `json:"eventTriggerData"`
	AggregatableTriggerData          []StorageAttributionReportingAggregatableTriggerData        `json:"aggregatableTriggerData"`
	AggregatableValues               []StorageAttributionReportingAggregatableValueEntry         `json:"aggregatableValues"`
	AggregatableFilteringIDMaxBytes  int                                                         `json:"aggregatableFilteringIdMaxBytes"`
	DebugReporting                   bool                                                        `json:"debugReporting"`
	AggregationCoordinatorOrigin     *string                                                     `json:"aggregationCoordinatorOrigin,omitempty"`
	SourceRegistrationTimeConfig     StorageAttributionReportingSourceRegistrationTimeConfig     `json:"sourceRegistrationTimeConfig"`
	TriggerContextID                 *string                                                     `json:"triggerContextId,omitempty"`
	AggregatableDebugReportingConfig StorageAttributionReportingAggregatableDebugReportingConfig `json:"aggregatableDebugReportingConfig"`
	Scopes                           []string                                                    `json:"scopes"`
	NamedBudgets                     []StorageAttributionReportingNamedBudgetCandidate           `json:"namedBudgets"`
}

type StorageAttributionReportingEventLevelResult string

type StorageAttributionReportingAggregatableResult string

type StorageAttributionReportingReportResult string

type StorageRelatedWebsiteSet struct {
	// The primary site of this set, along with the ccTLDs if there is any.
	PrimarySites []string `json:"primarySites"`
	// The associated sites of this set, along with the ccTLDs if there is any.
	AssociatedSites []string `json:"associatedSites"`
	// The service sites of this set, along with the ccTLDs if there is any.
	ServiceSites []string `json:"serviceSites"`
}

type StorageGetStorageKeyForFrameParams struct {
	SessionID string      `json:"-"`
	FrameID   PageFrameID `json:"frameId"`
}

type StorageGetStorageKeyForFrameResult struct {
	StorageKey StorageSerializedStorageKey `json:"storageKey"`
}

type StorageGetStorageKeyParams struct {
	SessionID string       `json:"-"`
	FrameID   *PageFrameID `json:"frameId,omitempty"`
}

type StorageGetStorageKeyResult struct {
	StorageKey StorageSerializedStorageKey `json:"storageKey"`
}

type StorageClearDataForOriginParams struct {
	SessionID string `json:"-"`
	// Security origin.
	Origin string `json:"origin"`
	// Comma separated list of StorageType to clear.
	StorageTypes string `json:"storageTypes"`
}

type StorageClearDataForOriginResult struct {
}

type StorageClearDataForStorageKeyParams struct {
	SessionID string `json:"-"`
	// Storage key.
	StorageKey string `json:"storageKey"`
	// Comma separated list of StorageType to clear.
	StorageTypes string `json:"storageTypes"`
}

type StorageClearDataForStorageKeyResult struct {
}

type StorageGetCookiesParams struct {
	SessionID string `json:"-"`
	// Browser context to use when called on the browser endpoint.
	BrowserContextID *BrowserBrowserContextID `json:"browserContextId,omitempty"`
}

type StorageGetCookiesResult struct {
	// Array of cookie objects.
	Cookies []NetworkCookie `json:"cookies"`
}

type StorageSetCookiesParams struct {
	SessionID string `json:"-"`
	// Cookies to be set.
	Cookies []NetworkCookieParam `json:"cookies"`
	// Browser context to use when called on the browser endpoint.
	BrowserContextID *BrowserBrowserContextID `json:"browserContextId,omitempty"`
}

type StorageSetCookiesResult struct {
}

type StorageClearCookiesParams struct {
	SessionID string `json:"-"`
	// Browser context to use when called on the browser endpoint.
	BrowserContextID *BrowserBrowserContextID `json:"browserContextId,omitempty"`
}

type StorageClearCookiesResult struct {
}

type StorageGetUsageAndQuotaParams struct {
	SessionID string `json:"-"`
	// Security origin.
	Origin string `json:"origin"`
}

type StorageGetUsageAndQuotaResult struct {
	// Storage usage (bytes).
	Usage float64 `json:"usage"`
	// Storage quota (bytes).
	Quota float64 `json:"quota"`
	// Whether or not the origin has an active storage quota override
	OverrideActive bool `json:"overrideActive"`
	// Storage usage per type (bytes).
	UsageBreakdown []StorageUsageForType `json:"usageBreakdown"`
}

type StorageOverrideQuotaForOriginParams struct {
	SessionID string `json:"-"`
	// Security origin.
	Origin string `json:"origin"`
	// The quota size (in bytes) to override the original quota with.
	QuotaSize *float64 `json:"quotaSize,omitempty"`
}

type StorageOverrideQuotaForOriginResult struct {
}

type StorageTrackCacheStorageForOriginParams struct {
	SessionID string `json:"-"`
	// Security origin.
	Origin string `json:"origin"`
}

type StorageTrackCacheStorageForOriginResult struct {
}

type StorageTrackCacheStorageForStorageKeyParams struct {
	SessionID string `json:"-"`
	// Storage key.
	StorageKey string `json:"storageKey"`
}

type StorageTrackCacheStorageForStorageKeyResult struct {
}

type StorageTrackIndexedDBForOriginParams struct {
	SessionID string `json:"-"`
	// Security origin.
	Origin string `json:"origin"`
}

type StorageTrackIndexedDBForOriginResult struct {
}

type StorageTrackIndexedDBForStorageKeyParams struct {
	SessionID string `json:"-"`
	// Storage key.
	StorageKey string `json:"storageKey"`
}

type StorageTrackIndexedDBForStorageKeyResult struct {
}

type StorageUntrackCacheStorageForOriginParams struct {
	SessionID string `json:"-"`
	// Security origin.
	Origin string `json:"origin"`
}

type StorageUntrackCacheStorageForOriginResult struct {
}

type StorageUntrackCacheStorageForStorageKeyParams struct {
	SessionID string `json:"-"`
	// Storage key.
	StorageKey string `json:"storageKey"`
}

type StorageUntrackCacheStorageForStorageKeyResult struct {
}

type StorageUntrackIndexedDBForOriginParams struct {
	SessionID string `json:"-"`
	// Security origin.
	Origin string `json:"origin"`
}

type StorageUntrackIndexedDBForOriginResult struct {
}

type StorageUntrackIndexedDBForStorageKeyParams struct {
	SessionID string `json:"-"`
	// Storage key.
	StorageKey string `json:"storageKey"`
}

type StorageUntrackIndexedDBForStorageKeyResult struct {
}

type StorageGetTrustTokensParams struct {
	SessionID string `json:"-"`
}

type StorageGetTrustTokensResult struct {
	Tokens []StorageTrustTokens `json:"tokens"`
}

type StorageClearTrustTokensParams struct {
	SessionID    string `json:"-"`
	IssuerOrigin string `json:"issuerOrigin"`
}

type StorageClearTrustTokensResult struct {
	// True if any tokens were deleted, false otherwise.
	DidDeleteTokens bool `json:"didDeleteTokens"`
}

type StorageGetInterestGroupDetailsParams struct {
	SessionID   string `json:"-"`
	OwnerOrigin string `json:"ownerOrigin"`
	Name        string `json:"name"`
}

type StorageGetInterestGroupDetailsResult struct {
	// This largely corresponds to:
	Details map[string]any `json:"details"`
}

type StorageSetInterestGroupTrackingParams struct {
	SessionID string `json:"-"`
	Enable    bool   `json:"enable"`
}

type StorageSetInterestGroupTrackingResult struct {
}

type StorageSetInterestGroupAuctionTrackingParams struct {
	SessionID string `json:"-"`
	Enable    bool   `json:"enable"`
}

type StorageSetInterestGroupAuctionTrackingResult struct {
}

type StorageGetSharedStorageMetadataParams struct {
	SessionID   string `json:"-"`
	OwnerOrigin string `json:"ownerOrigin"`
}

type StorageGetSharedStorageMetadataResult struct {
	Metadata StorageSharedStorageMetadata `json:"metadata"`
}

type StorageGetSharedStorageEntriesParams struct {
	SessionID   string `json:"-"`
	OwnerOrigin string `json:"ownerOrigin"`
}

type StorageGetSharedStorageEntriesResult struct {
	Entries []StorageSharedStorageEntry `json:"entries"`
}

type StorageSetSharedStorageEntryParams struct {
	SessionID   string `json:"-"`
	OwnerOrigin string `json:"ownerOrigin"`
	Key         string `json:"key"`
	Value       string `json:"value"`
	// If `ignoreIfPresent` is included and true, then only sets the entry if
	IgnoreIfPresent *bool `json:"ignoreIfPresent,omitempty"`
}

type StorageSetSharedStorageEntryResult struct {
}

type StorageDeleteSharedStorageEntryParams struct {
	SessionID   string `json:"-"`
	OwnerOrigin string `json:"ownerOrigin"`
	Key         string `json:"key"`
}

type StorageDeleteSharedStorageEntryResult struct {
}

type StorageClearSharedStorageEntriesParams struct {
	SessionID   string `json:"-"`
	OwnerOrigin string `json:"ownerOrigin"`
}

type StorageClearSharedStorageEntriesResult struct {
}

type StorageResetSharedStorageBudgetParams struct {
	SessionID   string `json:"-"`
	OwnerOrigin string `json:"ownerOrigin"`
}

type StorageResetSharedStorageBudgetResult struct {
}

type StorageSetSharedStorageTrackingParams struct {
	SessionID string `json:"-"`
	Enable    bool   `json:"enable"`
}

type StorageSetSharedStorageTrackingResult struct {
}

type StorageSetStorageBucketTrackingParams struct {
	SessionID  string `json:"-"`
	StorageKey string `json:"storageKey"`
	Enable     bool   `json:"enable"`
}

type StorageSetStorageBucketTrackingResult struct {
}

type StorageDeleteStorageBucketParams struct {
	SessionID string               `json:"-"`
	Bucket    StorageStorageBucket `json:"bucket"`
}

type StorageDeleteStorageBucketResult struct {
}

type StorageRunBounceTrackingMitigationsParams struct {
	SessionID string `json:"-"`
}

type StorageRunBounceTrackingMitigationsResult struct {
	DeletedSites []string `json:"deletedSites"`
}

type StorageSetAttributionReportingLocalTestingModeParams struct {
	SessionID string `json:"-"`
	// If enabled, noise is suppressed and reports are sent immediately.
	Enabled bool `json:"enabled"`
}

type StorageSetAttributionReportingLocalTestingModeResult struct {
}

type StorageSetAttributionReportingTrackingParams struct {
	SessionID string `json:"-"`
	Enable    bool   `json:"enable"`
}

type StorageSetAttributionReportingTrackingResult struct {
}

type StorageSendPendingAttributionReportsParams struct {
	SessionID string `json:"-"`
}

type StorageSendPendingAttributionReportsResult struct {
	// The number of reports that were sent.
	NumSent int `json:"numSent"`
}

type StorageGetRelatedWebsiteSetsParams struct {
	SessionID string `json:"-"`
}

type StorageGetRelatedWebsiteSetsResult struct {
	Sets []StorageRelatedWebsiteSet `json:"sets"`
}

type StorageGetAffectedUrlsForThirdPartyCookieMetadataParams struct {
	SessionID string `json:"-"`
	// The URL of the page currently being visited.
	FirstPartyURL string `json:"firstPartyUrl"`
	// The list of embedded resource URLs from the page.
	ThirdPartyUrls []string `json:"thirdPartyUrls"`
}

type StorageGetAffectedUrlsForThirdPartyCookieMetadataResult struct {
	// Array of matching URLs. If there is a primary pattern match for the first-
	MatchedUrls []string `json:"matchedUrls"`
}

type StorageSetProtectedAudienceKAnonymityParams struct {
	SessionID string   `json:"-"`
	Owner     string   `json:"owner"`
	Name      string   `json:"name"`
	Hashes    []string `json:"hashes"`
}

type StorageSetProtectedAudienceKAnonymityResult struct {
}

type StorageCacheStorageContentUpdatedEvent struct {
	// Origin to update.
	Origin string `json:"origin"`
	// Storage key to update.
	StorageKey string `json:"storageKey"`
	// Storage bucket to update.
	BucketID string `json:"bucketId"`
	// Name of cache in origin.
	CacheName string `json:"cacheName"`
}

type StorageCacheStorageListUpdatedEvent struct {
	// Origin to update.
	Origin string `json:"origin"`
	// Storage key to update.
	StorageKey string `json:"storageKey"`
	// Storage bucket to update.
	BucketID string `json:"bucketId"`
}

type StorageIndexedDBContentUpdatedEvent struct {
	// Origin to update.
	Origin string `json:"origin"`
	// Storage key to update.
	StorageKey string `json:"storageKey"`
	// Storage bucket to update.
	BucketID string `json:"bucketId"`
	// Database to update.
	DatabaseName string `json:"databaseName"`
	// ObjectStore to update.
	ObjectStoreName string `json:"objectStoreName"`
}

type StorageIndexedDBListUpdatedEvent struct {
	// Origin to update.
	Origin string `json:"origin"`
	// Storage key to update.
	StorageKey string `json:"storageKey"`
	// Storage bucket to update.
	BucketID string `json:"bucketId"`
}

type StorageInterestGroupAccessedEvent struct {
	AccessTime  NetworkTimeSinceEpoch          `json:"accessTime"`
	Type        StorageInterestGroupAccessType `json:"type"`
	OwnerOrigin string                         `json:"ownerOrigin"`
	Name        string                         `json:"name"`
	// For topLevelBid/topLevelAdditionalBid, and when appropriate,
	ComponentSellerOrigin *string `json:"componentSellerOrigin,omitempty"`
	// For bid or somethingBid event, if done locally and not on a server.
	Bid         *float64 `json:"bid,omitempty"`
	BidCurrency *string  `json:"bidCurrency,omitempty"`
	// For non-global events --- links to interestGroupAuctionEvent
	UniqueAuctionID *StorageInterestGroupAuctionID `json:"uniqueAuctionId,omitempty"`
}

type StorageInterestGroupAuctionEventOccurredEvent struct {
	EventTime       NetworkTimeSinceEpoch                `json:"eventTime"`
	Type            StorageInterestGroupAuctionEventType `json:"type"`
	UniqueAuctionID StorageInterestGroupAuctionID        `json:"uniqueAuctionId"`
	// Set for child auctions.
	ParentAuctionID *StorageInterestGroupAuctionID `json:"parentAuctionId,omitempty"`
	// Set for started and configResolved
	AuctionConfig map[string]any `json:"auctionConfig,omitempty"`
}

type StorageInterestGroupAuctionNetworkRequestCreatedEvent struct {
	Type      StorageInterestGroupAuctionFetchType `json:"type"`
	RequestID NetworkRequestID                     `json:"requestId"`
	// This is the set of the auctions using the worklet that issued this
	Auctions []StorageInterestGroupAuctionID `json:"auctions"`
}

type StorageSharedStorageAccessedEvent struct {
	// Time of the access.
	AccessTime NetworkTimeSinceEpoch `json:"accessTime"`
	// Enum value indicating the access scope.
	Scope StorageSharedStorageAccessScope `json:"scope"`
	// Enum value indicating the Shared Storage API method invoked.
	Method StorageSharedStorageAccessMethod `json:"method"`
	// DevTools Frame Token for the primary frame tree's root.
	MainFrameID PageFrameID `json:"mainFrameId"`
	// Serialization of the origin owning the Shared Storage data.
	OwnerOrigin string `json:"ownerOrigin"`
	// Serialization of the site owning the Shared Storage data.
	OwnerSite string `json:"ownerSite"`
	// The sub-parameters wrapped by `params` are all optional and their
	Params StorageSharedStorageAccessParams `json:"params"`
}

type StorageSharedStorageWorkletOperationExecutionFinishedEvent struct {
	// Time that the operation finished.
	FinishedTime NetworkTimeSinceEpoch `json:"finishedTime"`
	// Time, in microseconds, from start of shared storage JS API call until
	ExecutionTime int `json:"executionTime"`
	// Enum value indicating the Shared Storage API method invoked.
	Method StorageSharedStorageAccessMethod `json:"method"`
	// ID of the operation call.
	OperationID string `json:"operationId"`
	// Hex representation of the DevTools token used as the TargetID for the
	WorkletTargetID TargetTargetID `json:"workletTargetId"`
	// DevTools Frame Token for the primary frame tree's root.
	MainFrameID PageFrameID `json:"mainFrameId"`
	// Serialization of the origin owning the Shared Storage data.
	OwnerOrigin string `json:"ownerOrigin"`
}

type StorageStorageBucketCreatedOrUpdatedEvent struct {
	BucketInfo StorageStorageBucketInfo `json:"bucketInfo"`
}

type StorageStorageBucketDeletedEvent struct {
	BucketID string `json:"bucketId"`
}

type StorageAttributionReportingSourceRegisteredEvent struct {
	Registration StorageAttributionReportingSourceRegistration       `json:"registration"`
	Result       StorageAttributionReportingSourceRegistrationResult `json:"result"`
}

type StorageAttributionReportingTriggerRegisteredEvent struct {
	Registration StorageAttributionReportingTriggerRegistration `json:"registration"`
	EventLevel   StorageAttributionReportingEventLevelResult    `json:"eventLevel"`
	Aggregatable StorageAttributionReportingAggregatableResult  `json:"aggregatable"`
}

type StorageAttributionReportingReportSentEvent struct {
	URL    string                                  `json:"url"`
	Body   map[string]any                          `json:"body"`
	Result StorageAttributionReportingReportResult `json:"result"`
	// If result is `sent`, populated with net/HTTP status.
	NetError       *int    `json:"netError,omitempty"`
	NetErrorName   *string `json:"netErrorName,omitempty"`
	HTTPStatusCode *int    `json:"httpStatusCode,omitempty"`
}

type StorageAttributionReportingVerboseDebugReportSentEvent struct {
	URL            string           `json:"url"`
	Body           []map[string]any `json:"body,omitempty"`
	NetError       *int             `json:"netError,omitempty"`
	NetErrorName   *string          `json:"netErrorName,omitempty"`
	HTTPStatusCode *int             `json:"httpStatusCode,omitempty"`
}

type SystemInfoGPUDevice struct {
	// PCI ID of the GPU vendor, if available; 0 otherwise.
	VendorID float64 `json:"vendorId"`
	// PCI ID of the GPU device, if available; 0 otherwise.
	DeviceID float64 `json:"deviceId"`
	// Sub sys ID of the GPU, only available on Windows.
	SubSysID *float64 `json:"subSysId,omitempty"`
	// Revision of the GPU, only available on Windows.
	Revision *float64 `json:"revision,omitempty"`
	// String description of the GPU vendor, if the PCI ID is not available.
	VendorString string `json:"vendorString"`
	// String description of the GPU device, if the PCI ID is not available.
	DeviceString string `json:"deviceString"`
	// String description of the GPU driver vendor.
	DriverVendor string `json:"driverVendor"`
	// String description of the GPU driver version.
	DriverVersion string `json:"driverVersion"`
}

type SystemInfoSize struct {
	// Width in pixels.
	Width int `json:"width"`
	// Height in pixels.
	Height int `json:"height"`
}

type SystemInfoVideoDecodeAcceleratorCapability struct {
	// Video codec profile that is supported, e.g. VP9 Profile 2.
	Profile string `json:"profile"`
	// Maximum video dimensions in pixels supported for this |profile|.
	MaxResolution SystemInfoSize `json:"maxResolution"`
	// Minimum video dimensions in pixels supported for this |profile|.
	MinResolution SystemInfoSize `json:"minResolution"`
}

type SystemInfoVideoEncodeAcceleratorCapability struct {
	// Video codec profile that is supported, e.g H264 Main.
	Profile string `json:"profile"`
	// Maximum video dimensions in pixels supported for this |profile|.
	MaxResolution SystemInfoSize `json:"maxResolution"`
	// Maximum encoding framerate in frames per second supported for this
	MaxFramerateNumerator   int `json:"maxFramerateNumerator"`
	MaxFramerateDenominator int `json:"maxFramerateDenominator"`
}

type SystemInfoSubsamplingFormat string

type SystemInfoImageType string

type SystemInfoGPUInfo struct {
	// The graphics devices on the system. Element 0 is the primary GPU.
	Devices []SystemInfoGPUDevice `json:"devices"`
	// An optional dictionary of additional GPU related attributes.
	AuxAttributes map[string]any `json:"auxAttributes,omitempty"`
	// An optional dictionary of graphics features and their status.
	FeatureStatus map[string]any `json:"featureStatus,omitempty"`
	// An optional array of GPU driver bug workarounds.
	DriverBugWorkarounds []string `json:"driverBugWorkarounds"`
	// Supported accelerated video decoding capabilities.
	VideoDecoding []SystemInfoVideoDecodeAcceleratorCapability `json:"videoDecoding"`
	// Supported accelerated video encoding capabilities.
	VideoEncoding []SystemInfoVideoEncodeAcceleratorCapability `json:"videoEncoding"`
}

type SystemInfoProcessInfo struct {
	// Specifies process type.
	Type string `json:"type"`
	// Specifies process id.
	ID int `json:"id"`
	// Specifies cumulative CPU usage in seconds across all threads of the
	CPUTime float64 `json:"cpuTime"`
}

type SystemInfoGetInfoParams struct {
	SessionID string `json:"-"`
}

type SystemInfoGetInfoResult struct {
	// Information about the GPUs on the system.
	GPU SystemInfoGPUInfo `json:"gpu"`
	// A platform-dependent description of the model of the machine. On Mac OS, this is, for
	ModelName string `json:"modelName"`
	// A platform-dependent description of the version of the machine. On Mac OS, this is, for
	ModelVersion string `json:"modelVersion"`
	// The command line string used to launch the browser. Will be the empty string if not
	CommandLine string `json:"commandLine"`
}

type SystemInfoGetFeatureStateParams struct {
	SessionID    string `json:"-"`
	FeatureState string `json:"featureState"`
}

type SystemInfoGetFeatureStateResult struct {
	FeatureEnabled bool `json:"featureEnabled"`
}

type SystemInfoGetProcessInfoParams struct {
	SessionID string `json:"-"`
}

type SystemInfoGetProcessInfoResult struct {
	// An array of process info blocks.
	ProcessInfo []SystemInfoProcessInfo `json:"processInfo"`
}

type TargetTargetID string

type TargetSessionID string

type TargetTargetInfo struct {
	TargetID TargetTargetID `json:"targetId"`
	// List of types: https://source.chromium.org/chromium/chromium/src/+/main:content/browser/devtools/devtools_agent_host_impl.cc?ss=chromium&q=f:devtools%20-f:out%20%22::kTypeTab%5B%5D
	Type  string `json:"type"`
	Title string `json:"title"`
	URL   string `json:"url"`
	// Whether the target has an attached client.
	Attached bool `json:"attached"`
	// Opener target Id
	OpenerID *TargetTargetID `json:"openerId,omitempty"`
	// Whether the target has access to the originating window.
	CanAccessOpener bool `json:"canAccessOpener"`
	// Frame id of originating window (is only set if target has an opener).
	OpenerFrameID *PageFrameID `json:"openerFrameId,omitempty"`
	// Id of the parent frame, only present for the "iframe" targets.
	ParentFrameID    *PageFrameID             `json:"parentFrameId,omitempty"`
	BrowserContextID *BrowserBrowserContextID `json:"browserContextId,omitempty"`
	// Provides additional details for specific target types. For example, for
	Subtype *string `json:"subtype,omitempty"`
}

type TargetFilterEntry struct {
	// If set, causes exclusion of matching targets from the list.
	Exclude *bool `json:"exclude,omitempty"`
	// If not present, matches any type.
	Type *string `json:"type,omitempty"`
}

type TargetTargetFilter []TargetFilterEntry

type TargetRemoteLocation struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

type TargetWindowState string

type TargetActivateTargetParams struct {
	SessionID string         `json:"-"`
	TargetID  TargetTargetID `json:"targetId"`
}

type TargetActivateTargetResult struct {
}

type TargetAttachToTargetParams struct {
	SessionID string         `json:"-"`
	TargetID  TargetTargetID `json:"targetId"`
	// Enables "flat" access to the session via specifying sessionId attribute in the commands.
	Flatten *bool `json:"flatten,omitempty"`
}

type TargetAttachToTargetResult struct {
	// Id assigned to the session.
	SessionID TargetSessionID `json:"sessionId"`
}

type TargetAttachToBrowserTargetParams struct {
	SessionID string `json:"-"`
}

type TargetAttachToBrowserTargetResult struct {
	// Id assigned to the session.
	SessionID TargetSessionID `json:"sessionId"`
}

type TargetCloseTargetParams struct {
	SessionID string         `json:"-"`
	TargetID  TargetTargetID `json:"targetId"`
}

type TargetCloseTargetResult struct {
	// Always set to true. If an error occurs, the response indicates protocol error.
	Success bool `json:"success"`
}

type TargetExposeDevToolsProtocolParams struct {
	SessionID string         `json:"-"`
	TargetID  TargetTargetID `json:"targetId"`
	// Binding name, 'cdp' if not specified.
	BindingName *string `json:"bindingName,omitempty"`
	// If true, inherits the current root session's permissions (default: false).
	InheritPermissions *bool `json:"inheritPermissions,omitempty"`
}

type TargetExposeDevToolsProtocolResult struct {
}

type TargetCreateBrowserContextParams struct {
	SessionID string `json:"-"`
	// If specified, disposes this context when debugging session disconnects.
	DisposeOnDetach *bool `json:"disposeOnDetach,omitempty"`
	// Proxy server, similar to the one passed to --proxy-server
	ProxyServer *string `json:"proxyServer,omitempty"`
	// Proxy bypass list, similar to the one passed to --proxy-bypass-list
	ProxyBypassList *string `json:"proxyBypassList,omitempty"`
	// An optional list of origins to grant unlimited cross-origin access to.
	OriginsWithUniversalNetworkAccess []string `json:"originsWithUniversalNetworkAccess,omitempty"`
}

type TargetCreateBrowserContextResult struct {
	// The id of the context created.
	BrowserContextID BrowserBrowserContextID `json:"browserContextId"`
}

type TargetGetBrowserContextsParams struct {
	SessionID string `json:"-"`
}

type TargetGetBrowserContextsResult struct {
	// An array of browser context ids.
	BrowserContextIds []BrowserBrowserContextID `json:"browserContextIds"`
	// The id of the default browser context if available.
	DefaultBrowserContextID *BrowserBrowserContextID `json:"defaultBrowserContextId,omitempty"`
}

type TargetCreateTargetParams struct {
	SessionID string `json:"-"`
	// The initial URL the page will be navigated to. An empty string indicates about:blank.
	URL string `json:"url"`
	// Frame left origin in DIP (requires newWindow to be true or headless shell).
	Left *int `json:"left,omitempty"`
	// Frame top origin in DIP (requires newWindow to be true or headless shell).
	Top *int `json:"top,omitempty"`
	// Frame width in DIP (requires newWindow to be true or headless shell).
	Width *int `json:"width,omitempty"`
	// Frame height in DIP (requires newWindow to be true or headless shell).
	Height *int `json:"height,omitempty"`
	// Frame window state (requires newWindow to be true or headless shell).
	WindowState *TargetWindowState `json:"windowState,omitempty"`
	// The browser context to create the page in.
	BrowserContextID *BrowserBrowserContextID `json:"browserContextId,omitempty"`
	// Whether BeginFrames for this target will be controlled via DevTools (headless shell only,
	EnableBeginFrameControl *bool `json:"enableBeginFrameControl,omitempty"`
	// Whether to create a new Window or Tab (false by default, not supported by headless shell).
	NewWindow *bool `json:"newWindow,omitempty"`
	// Whether to create the target in background or foreground (false by default, not supported
	Background *bool `json:"background,omitempty"`
	// Whether to create the target of type "tab".
	ForTab *bool `json:"forTab,omitempty"`
	// Whether to create a hidden target. The hidden target is observable via protocol, but not
	Hidden *bool `json:"hidden,omitempty"`
	// If specified, the option is used to determine if the new target should
	Focus *bool `json:"focus,omitempty"`
}

type TargetCreateTargetResult struct {
	// The id of the page opened.
	TargetID TargetTargetID `json:"targetId"`
}

type TargetDetachFromTargetParams struct {
	SessionID string `json:"-"`
	// Session to detach.
	SessionIDValue *TargetSessionID `json:"sessionId,omitempty"`
	// Deprecated.
	TargetID *TargetTargetID `json:"targetId,omitempty"`
}

type TargetDetachFromTargetResult struct {
}

type TargetDisposeBrowserContextParams struct {
	SessionID        string                  `json:"-"`
	BrowserContextID BrowserBrowserContextID `json:"browserContextId"`
}

type TargetDisposeBrowserContextResult struct {
}

type TargetGetTargetInfoParams struct {
	SessionID string          `json:"-"`
	TargetID  *TargetTargetID `json:"targetId,omitempty"`
}

type TargetGetTargetInfoResult struct {
	TargetInfo TargetTargetInfo `json:"targetInfo"`
}

type TargetGetTargetsParams struct {
	SessionID string `json:"-"`
	// Only targets matching filter will be reported. If filter is not specified
	Filter *TargetTargetFilter `json:"filter,omitempty"`
}

type TargetGetTargetsResult struct {
	// The list of targets.
	TargetInfos []TargetTargetInfo `json:"targetInfos"`
}

type TargetSendMessageToTargetParams struct {
	SessionID string `json:"-"`
	Message   string `json:"message"`
	// Identifier of the session.
	SessionIDValue *TargetSessionID `json:"sessionId,omitempty"`
	// Deprecated.
	TargetID *TargetTargetID `json:"targetId,omitempty"`
}

type TargetSendMessageToTargetResult struct {
}

type TargetSetAutoAttachParams struct {
	SessionID string `json:"-"`
	// Whether to auto-attach to related targets.
	AutoAttach bool `json:"autoAttach"`
	// Whether to pause new targets when attaching to them. Use `Runtime.runIfWaitingForDebugger`
	WaitForDebuggerOnStart bool `json:"waitForDebuggerOnStart"`
	// Enables "flat" access to the session via specifying sessionId attribute in the commands.
	Flatten *bool `json:"flatten,omitempty"`
	// Only targets matching filter will be attached.
	Filter *TargetTargetFilter `json:"filter,omitempty"`
}

type TargetSetAutoAttachResult struct {
}

type TargetAutoAttachRelatedParams struct {
	SessionID string         `json:"-"`
	TargetID  TargetTargetID `json:"targetId"`
	// Whether to pause new targets when attaching to them. Use `Runtime.runIfWaitingForDebugger`
	WaitForDebuggerOnStart bool `json:"waitForDebuggerOnStart"`
	// Only targets matching filter will be attached.
	Filter *TargetTargetFilter `json:"filter,omitempty"`
}

type TargetAutoAttachRelatedResult struct {
}

type TargetSetDiscoverTargetsParams struct {
	SessionID string `json:"-"`
	// Whether to discover available targets.
	Discover bool `json:"discover"`
	// Only targets matching filter will be attached. If `discover` is false,
	Filter *TargetTargetFilter `json:"filter,omitempty"`
}

type TargetSetDiscoverTargetsResult struct {
}

type TargetSetRemoteLocationsParams struct {
	SessionID string `json:"-"`
	// List of remote locations.
	Locations []TargetRemoteLocation `json:"locations"`
}

type TargetSetRemoteLocationsResult struct {
}

type TargetGetDevToolsTargetParams struct {
	SessionID string `json:"-"`
	// Page or tab target ID.
	TargetID TargetTargetID `json:"targetId"`
}

type TargetGetDevToolsTargetResult struct {
	// The targetId of DevTools page target if exists.
	TargetID *TargetTargetID `json:"targetId,omitempty"`
}

type TargetOpenDevToolsParams struct {
	SessionID string `json:"-"`
	// This can be the page or tab target ID.
	TargetID TargetTargetID `json:"targetId"`
	// The id of the panel we want DevTools to open initially. Currently
	PanelID *string `json:"panelId,omitempty"`
}

type TargetOpenDevToolsResult struct {
	// The targetId of DevTools page target.
	TargetID TargetTargetID `json:"targetId"`
}

type TargetAttachedToTargetEvent struct {
	// Identifier assigned to the session used to send/receive messages.
	SessionID          TargetSessionID  `json:"sessionId"`
	TargetInfo         TargetTargetInfo `json:"targetInfo"`
	WaitingForDebugger bool             `json:"waitingForDebugger"`
}

type TargetDetachedFromTargetEvent struct {
	// Detached session identifier.
	SessionID TargetSessionID `json:"sessionId"`
	// Deprecated.
	TargetID *TargetTargetID `json:"targetId,omitempty"`
}

type TargetReceivedMessageFromTargetEvent struct {
	// Identifier of a session which sends a message.
	SessionID TargetSessionID `json:"sessionId"`
	Message   string          `json:"message"`
	// Deprecated.
	TargetID *TargetTargetID `json:"targetId,omitempty"`
}

type TargetTargetCreatedEvent struct {
	TargetInfo TargetTargetInfo `json:"targetInfo"`
}

func (e TargetTargetCreatedEvent) TargetID() string { return string(e.TargetInfo.TargetID) }

type TargetTargetDestroyedEvent struct {
	TargetID TargetTargetID `json:"targetId"`
}

type TargetTargetCrashedEvent struct {
	TargetID TargetTargetID `json:"targetId"`
	// Termination status type.
	Status string `json:"status"`
	// Termination error code.
	ErrorCode int `json:"errorCode"`
}

type TargetTargetInfoChangedEvent struct {
	TargetInfo TargetTargetInfo `json:"targetInfo"`
}

func (e TargetTargetInfoChangedEvent) TargetID() string { return string(e.TargetInfo.TargetID) }

type TetheringBindParams struct {
	SessionID string `json:"-"`
	// Port number to bind.
	Port int `json:"port"`
}

type TetheringBindResult struct {
}

type TetheringUnbindParams struct {
	SessionID string `json:"-"`
	// Port number to unbind.
	Port int `json:"port"`
}

type TetheringUnbindResult struct {
}

type TetheringAcceptedEvent struct {
	// Port number that was successfully bound.
	Port int `json:"port"`
	// Connection id to be used.
	ConnectionID string `json:"connectionId"`
}

type TracingMemoryDumpConfig map[string]any

type TracingTraceConfig struct {
	// Controls how the trace buffer stores data. The default is `recordUntilFull`.
	RecordMode *string `json:"recordMode,omitempty"`
	// Size of the trace buffer in kilobytes. If not specified or zero is passed, a default value
	TraceBufferSizeInKb *float64 `json:"traceBufferSizeInKb,omitempty"`
	// Turns on JavaScript stack sampling.
	EnableSampling *bool `json:"enableSampling,omitempty"`
	// Turns on system tracing.
	EnableSystrace *bool `json:"enableSystrace,omitempty"`
	// Turns on argument filter.
	EnableArgumentFilter *bool `json:"enableArgumentFilter,omitempty"`
	// Included category filters.
	IncludedCategories []string `json:"includedCategories,omitempty"`
	// Excluded category filters.
	ExcludedCategories []string `json:"excludedCategories,omitempty"`
	// Configuration to synthesize the delays in tracing.
	SyntheticDelays []string `json:"syntheticDelays,omitempty"`
	// Configuration for memory dump triggers. Used only when "memory-infra" category is enabled.
	MemoryDumpConfig *TracingMemoryDumpConfig `json:"memoryDumpConfig,omitempty"`
}

type TracingStreamFormat string

type TracingStreamCompression string

type TracingMemoryDumpLevelOfDetail string

type TracingTracingBackend string

type TracingEndParams struct {
	SessionID string `json:"-"`
}

type TracingEndResult struct {
}

type TracingGetCategoriesParams struct {
	SessionID string `json:"-"`
}

type TracingGetCategoriesResult struct {
	// A list of supported tracing categories.
	Categories []string `json:"categories"`
}

type TracingGetTrackEventDescriptorParams struct {
	SessionID string `json:"-"`
}

type TracingGetTrackEventDescriptorResult struct {
	// Base64-encoded serialized perfetto.protos.TrackEventDescriptor protobuf message. (Encoded as a base64 string when passed over JSON)
	Descriptor string `json:"descriptor"`
}

type TracingRecordClockSyncMarkerParams struct {
	SessionID string `json:"-"`
	// The ID of this clock sync marker
	SyncID string `json:"syncId"`
}

type TracingRecordClockSyncMarkerResult struct {
}

type TracingRequestMemoryDumpParams struct {
	SessionID string `json:"-"`
	// Enables more deterministic results by forcing garbage collection
	Deterministic *bool `json:"deterministic,omitempty"`
	// Specifies level of details in memory dump. Defaults to "detailed".
	LevelOfDetail *TracingMemoryDumpLevelOfDetail `json:"levelOfDetail,omitempty"`
}

type TracingRequestMemoryDumpResult struct {
	// GUID of the resulting global memory dump.
	DumpGuid string `json:"dumpGuid"`
	// True iff the global memory dump succeeded.
	Success bool `json:"success"`
}

type TracingStartParams struct {
	SessionID string `json:"-"`
	// Category/tag filter
	Categories *string `json:"categories,omitempty"`
	// Tracing options
	Options *string `json:"options,omitempty"`
	// If set, the agent will issue bufferUsage events at this interval, specified in milliseconds
	BufferUsageReportingInterval *float64 `json:"bufferUsageReportingInterval,omitempty"`
	// Whether to report trace events as series of dataCollected events or to save trace to a
	TransferMode *string `json:"transferMode,omitempty"`
	// Trace data format to use. This only applies when using `ReturnAsStream`
	StreamFormat *TracingStreamFormat `json:"streamFormat,omitempty"`
	// Compression format to use. This only applies when using `ReturnAsStream`
	StreamCompression *TracingStreamCompression `json:"streamCompression,omitempty"`
	TraceConfig       *TracingTraceConfig       `json:"traceConfig,omitempty"`
	// Base64-encoded serialized perfetto.protos.TraceConfig protobuf message
	PerfettoConfig *string `json:"perfettoConfig,omitempty"`
	// Backend type (defaults to `auto`)
	TracingBackend *TracingTracingBackend `json:"tracingBackend,omitempty"`
}

type TracingStartResult struct {
}

type TracingBufferUsageEvent struct {
	// A number in range [0..1] that indicates the used size of event buffer as a fraction of its
	PercentFull *float64 `json:"percentFull,omitempty"`
	// An approximate number of events in the trace log.
	EventCount *float64 `json:"eventCount,omitempty"`
	// A number in range [0..1] that indicates the used size of event buffer as a fraction of its
	Value *float64 `json:"value,omitempty"`
}

type TracingDataCollectedEvent struct {
	Value []map[string]any `json:"value"`
}

type TracingTracingCompleteEvent struct {
	// Indicates whether some trace data is known to have been lost, e.g. because the trace ring
	DataLossOccurred bool `json:"dataLossOccurred"`
	// A handle of the stream that holds resulting trace data.
	Stream *IOStreamHandle `json:"stream,omitempty"`
	// Trace data format of returned stream.
	TraceFormat *TracingStreamFormat `json:"traceFormat,omitempty"`
	// Compression format of returned stream.
	StreamCompression *TracingStreamCompression `json:"streamCompression,omitempty"`
}

type WebAudioGraphObjectID string

type WebAudioContextType string

type WebAudioContextState string

type WebAudioNodeType string

type WebAudioChannelCountMode string

type WebAudioChannelInterpretation string

type WebAudioParamType string

type WebAudioAutomationRate string

type WebAudioContextRealtimeData struct {
	// The current context time in second in BaseAudioContext.
	CurrentTime float64 `json:"currentTime"`
	// The time spent on rendering graph divided by render quantum duration,
	RenderCapacity float64 `json:"renderCapacity"`
	// A running mean of callback interval.
	CallbackIntervalMean float64 `json:"callbackIntervalMean"`
	// A running variance of callback interval.
	CallbackIntervalVariance float64 `json:"callbackIntervalVariance"`
}

type WebAudioBaseAudioContext struct {
	ContextID    WebAudioGraphObjectID        `json:"contextId"`
	ContextType  WebAudioContextType          `json:"contextType"`
	ContextState WebAudioContextState         `json:"contextState"`
	RealtimeData *WebAudioContextRealtimeData `json:"realtimeData,omitempty"`
	// Platform-dependent callback buffer size.
	CallbackBufferSize float64 `json:"callbackBufferSize"`
	// Number of output channels supported by audio hardware in use.
	MaxOutputChannelCount float64 `json:"maxOutputChannelCount"`
	// Context sample rate.
	SampleRate float64 `json:"sampleRate"`
}

type WebAudioAudioListener struct {
	ListenerID WebAudioGraphObjectID `json:"listenerId"`
	ContextID  WebAudioGraphObjectID `json:"contextId"`
}

type WebAudioAudioNode struct {
	NodeID                WebAudioGraphObjectID         `json:"nodeId"`
	ContextID             WebAudioGraphObjectID         `json:"contextId"`
	NodeType              WebAudioNodeType              `json:"nodeType"`
	NumberOfInputs        float64                       `json:"numberOfInputs"`
	NumberOfOutputs       float64                       `json:"numberOfOutputs"`
	ChannelCount          float64                       `json:"channelCount"`
	ChannelCountMode      WebAudioChannelCountMode      `json:"channelCountMode"`
	ChannelInterpretation WebAudioChannelInterpretation `json:"channelInterpretation"`
}

type WebAudioAudioParam struct {
	ParamID      WebAudioGraphObjectID  `json:"paramId"`
	NodeID       WebAudioGraphObjectID  `json:"nodeId"`
	ContextID    WebAudioGraphObjectID  `json:"contextId"`
	ParamType    WebAudioParamType      `json:"paramType"`
	Rate         WebAudioAutomationRate `json:"rate"`
	DefaultValue float64                `json:"defaultValue"`
	MinValue     float64                `json:"minValue"`
	MaxValue     float64                `json:"maxValue"`
}

type WebAudioEnableParams struct {
	SessionID string `json:"-"`
}

type WebAudioEnableResult struct {
}

type WebAudioDisableParams struct {
	SessionID string `json:"-"`
}

type WebAudioDisableResult struct {
}

type WebAudioGetRealtimeDataParams struct {
	SessionID string                `json:"-"`
	ContextID WebAudioGraphObjectID `json:"contextId"`
}

type WebAudioGetRealtimeDataResult struct {
	RealtimeData WebAudioContextRealtimeData `json:"realtimeData"`
}

type WebAudioContextCreatedEvent struct {
	Context WebAudioBaseAudioContext `json:"context"`
}

type WebAudioContextWillBeDestroyedEvent struct {
	ContextID WebAudioGraphObjectID `json:"contextId"`
}

type WebAudioContextChangedEvent struct {
	Context WebAudioBaseAudioContext `json:"context"`
}

type WebAudioAudioListenerCreatedEvent struct {
	Listener WebAudioAudioListener `json:"listener"`
}

type WebAudioAudioListenerWillBeDestroyedEvent struct {
	ContextID  WebAudioGraphObjectID `json:"contextId"`
	ListenerID WebAudioGraphObjectID `json:"listenerId"`
}

type WebAudioAudioNodeCreatedEvent struct {
	Node WebAudioAudioNode `json:"node"`
}

type WebAudioAudioNodeWillBeDestroyedEvent struct {
	ContextID WebAudioGraphObjectID `json:"contextId"`
	NodeID    WebAudioGraphObjectID `json:"nodeId"`
}

type WebAudioAudioParamCreatedEvent struct {
	Param WebAudioAudioParam `json:"param"`
}

type WebAudioAudioParamWillBeDestroyedEvent struct {
	ContextID WebAudioGraphObjectID `json:"contextId"`
	NodeID    WebAudioGraphObjectID `json:"nodeId"`
	ParamID   WebAudioGraphObjectID `json:"paramId"`
}

type WebAudioNodesConnectedEvent struct {
	ContextID             WebAudioGraphObjectID `json:"contextId"`
	SourceID              WebAudioGraphObjectID `json:"sourceId"`
	DestinationID         WebAudioGraphObjectID `json:"destinationId"`
	SourceOutputIndex     *float64              `json:"sourceOutputIndex,omitempty"`
	DestinationInputIndex *float64              `json:"destinationInputIndex,omitempty"`
}

type WebAudioNodesDisconnectedEvent struct {
	ContextID             WebAudioGraphObjectID `json:"contextId"`
	SourceID              WebAudioGraphObjectID `json:"sourceId"`
	DestinationID         WebAudioGraphObjectID `json:"destinationId"`
	SourceOutputIndex     *float64              `json:"sourceOutputIndex,omitempty"`
	DestinationInputIndex *float64              `json:"destinationInputIndex,omitempty"`
}

type WebAudioNodeParamConnectedEvent struct {
	ContextID         WebAudioGraphObjectID `json:"contextId"`
	SourceID          WebAudioGraphObjectID `json:"sourceId"`
	DestinationID     WebAudioGraphObjectID `json:"destinationId"`
	SourceOutputIndex *float64              `json:"sourceOutputIndex,omitempty"`
}

type WebAudioNodeParamDisconnectedEvent struct {
	ContextID         WebAudioGraphObjectID `json:"contextId"`
	SourceID          WebAudioGraphObjectID `json:"sourceId"`
	DestinationID     WebAudioGraphObjectID `json:"destinationId"`
	SourceOutputIndex *float64              `json:"sourceOutputIndex,omitempty"`
}

type WebAuthnAuthenticatorID string

type WebAuthnAuthenticatorProtocol string

type WebAuthnCtap2Version string

type WebAuthnAuthenticatorTransport string

type WebAuthnVirtualAuthenticatorOptions struct {
	Protocol WebAuthnAuthenticatorProtocol `json:"protocol"`
	// Defaults to ctap2_0. Ignored if |protocol| == u2f.
	Ctap2Version *WebAuthnCtap2Version          `json:"ctap2Version,omitempty"`
	Transport    WebAuthnAuthenticatorTransport `json:"transport"`
	// Defaults to false.
	HasResidentKey *bool `json:"hasResidentKey,omitempty"`
	// Defaults to false.
	HasUserVerification *bool `json:"hasUserVerification,omitempty"`
	// If set to true, the authenticator will support the largeBlob extension.
	HasLargeBlob *bool `json:"hasLargeBlob,omitempty"`
	// If set to true, the authenticator will support the credBlob extension.
	HasCredBlob *bool `json:"hasCredBlob,omitempty"`
	// If set to true, the authenticator will support the minPinLength extension.
	HasMinPinLength *bool `json:"hasMinPinLength,omitempty"`
	// If set to true, the authenticator will support the prf extension.
	HasPrf *bool `json:"hasPrf,omitempty"`
	// If set to true, tests of user presence will succeed immediately.
	AutomaticPresenceSimulation *bool `json:"automaticPresenceSimulation,omitempty"`
	// Sets whether User Verification succeeds or fails for an authenticator.
	IsUserVerified *bool `json:"isUserVerified,omitempty"`
	// Credentials created by this authenticator will have the backup
	DefaultBackupEligibility *bool `json:"defaultBackupEligibility,omitempty"`
	// Credentials created by this authenticator will have the backup state
	DefaultBackupState *bool `json:"defaultBackupState,omitempty"`
}

type WebAuthnCredential struct {
	CredentialID         string `json:"credentialId"`
	IsResidentCredential bool   `json:"isResidentCredential"`
	// Relying Party ID the credential is scoped to. Must be set when adding a
	RpID *string `json:"rpId,omitempty"`
	// The ECDSA P-256 private key in PKCS#8 format. (Encoded as a base64 string when passed over JSON)
	PrivateKey string `json:"privateKey"`
	// An opaque byte sequence with a maximum size of 64 bytes mapping the
	UserHandle *string `json:"userHandle,omitempty"`
	// Signature counter. This is incremented by one for each successful
	SignCount int `json:"signCount"`
	// The large blob associated with the credential.
	LargeBlob *string `json:"largeBlob,omitempty"`
	// Assertions returned by this credential will have the backup eligibility
	BackupEligibility *bool `json:"backupEligibility,omitempty"`
	// Assertions returned by this credential will have the backup state (BS)
	BackupState *bool `json:"backupState,omitempty"`
	// The credential's user.name property. Equivalent to empty if not set.
	UserName *string `json:"userName,omitempty"`
	// The credential's user.displayName property. Equivalent to empty if
	UserDisplayName *string `json:"userDisplayName,omitempty"`
}

type WebAuthnEnableParams struct {
	SessionID string `json:"-"`
	// Whether to enable the WebAuthn user interface. Enabling the UI is
	EnableUI *bool `json:"enableUI,omitempty"`
}

type WebAuthnEnableResult struct {
}

type WebAuthnDisableParams struct {
	SessionID string `json:"-"`
}

type WebAuthnDisableResult struct {
}

type WebAuthnAddVirtualAuthenticatorParams struct {
	SessionID string                              `json:"-"`
	Options   WebAuthnVirtualAuthenticatorOptions `json:"options"`
}

type WebAuthnAddVirtualAuthenticatorResult struct {
	AuthenticatorID WebAuthnAuthenticatorID `json:"authenticatorId"`
}

type WebAuthnSetResponseOverrideBitsParams struct {
	SessionID       string                  `json:"-"`
	AuthenticatorID WebAuthnAuthenticatorID `json:"authenticatorId"`
	// If isBogusSignature is set, overrides the signature in the authenticator response to be zero.
	IsBogusSignature *bool `json:"isBogusSignature,omitempty"`
	// If isBadUV is set, overrides the UV bit in the flags in the authenticator response to
	IsBadUV *bool `json:"isBadUV,omitempty"`
	// If isBadUP is set, overrides the UP bit in the flags in the authenticator response to
	IsBadUP *bool `json:"isBadUP,omitempty"`
}

type WebAuthnSetResponseOverrideBitsResult struct {
}

type WebAuthnRemoveVirtualAuthenticatorParams struct {
	SessionID       string                  `json:"-"`
	AuthenticatorID WebAuthnAuthenticatorID `json:"authenticatorId"`
}

type WebAuthnRemoveVirtualAuthenticatorResult struct {
}

type WebAuthnAddCredentialParams struct {
	SessionID       string                  `json:"-"`
	AuthenticatorID WebAuthnAuthenticatorID `json:"authenticatorId"`
	Credential      WebAuthnCredential      `json:"credential"`
}

type WebAuthnAddCredentialResult struct {
}

type WebAuthnGetCredentialParams struct {
	SessionID       string                  `json:"-"`
	AuthenticatorID WebAuthnAuthenticatorID `json:"authenticatorId"`
	CredentialID    string                  `json:"credentialId"`
}

type WebAuthnGetCredentialResult struct {
	Credential WebAuthnCredential `json:"credential"`
}

type WebAuthnGetCredentialsParams struct {
	SessionID       string                  `json:"-"`
	AuthenticatorID WebAuthnAuthenticatorID `json:"authenticatorId"`
}

type WebAuthnGetCredentialsResult struct {
	Credentials []WebAuthnCredential `json:"credentials"`
}

type WebAuthnRemoveCredentialParams struct {
	SessionID       string                  `json:"-"`
	AuthenticatorID WebAuthnAuthenticatorID `json:"authenticatorId"`
	CredentialID    string                  `json:"credentialId"`
}

type WebAuthnRemoveCredentialResult struct {
}

type WebAuthnClearCredentialsParams struct {
	SessionID       string                  `json:"-"`
	AuthenticatorID WebAuthnAuthenticatorID `json:"authenticatorId"`
}

type WebAuthnClearCredentialsResult struct {
}

type WebAuthnSetUserVerifiedParams struct {
	SessionID       string                  `json:"-"`
	AuthenticatorID WebAuthnAuthenticatorID `json:"authenticatorId"`
	IsUserVerified  bool                    `json:"isUserVerified"`
}

type WebAuthnSetUserVerifiedResult struct {
}

type WebAuthnSetAutomaticPresenceSimulationParams struct {
	SessionID       string                  `json:"-"`
	AuthenticatorID WebAuthnAuthenticatorID `json:"authenticatorId"`
	Enabled         bool                    `json:"enabled"`
}

type WebAuthnSetAutomaticPresenceSimulationResult struct {
}

type WebAuthnSetCredentialPropertiesParams struct {
	SessionID         string                  `json:"-"`
	AuthenticatorID   WebAuthnAuthenticatorID `json:"authenticatorId"`
	CredentialID      string                  `json:"credentialId"`
	BackupEligibility *bool                   `json:"backupEligibility,omitempty"`
	BackupState       *bool                   `json:"backupState,omitempty"`
}

type WebAuthnSetCredentialPropertiesResult struct {
}

type WebAuthnCredentialAddedEvent struct {
	AuthenticatorID WebAuthnAuthenticatorID `json:"authenticatorId"`
	Credential      WebAuthnCredential      `json:"credential"`
}

type WebAuthnCredentialDeletedEvent struct {
	AuthenticatorID WebAuthnAuthenticatorID `json:"authenticatorId"`
	CredentialID    string                  `json:"credentialId"`
}

type WebAuthnCredentialUpdatedEvent struct {
	AuthenticatorID WebAuthnAuthenticatorID `json:"authenticatorId"`
	Credential      WebAuthnCredential      `json:"credential"`
}

type WebAuthnCredentialAssertedEvent struct {
	AuthenticatorID WebAuthnAuthenticatorID `json:"authenticatorId"`
	Credential      WebAuthnCredential      `json:"credential"`
}
