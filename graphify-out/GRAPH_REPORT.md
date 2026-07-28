# Graph Report - tv  (2026-07-29)

## Corpus Check
- 62 files · ~56,820 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 1716 nodes · 4557 edges · 107 communities (93 shown, 14 thin omitted)
- Extraction: 84% EXTRACTED · 16% INFERRED · 0% AMBIGUOUS · INFERRED: 708 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `3d1c3eed`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- [[_COMMUNITY_Community 0|Community 0]]
- [[_COMMUNITY_Community 1|Community 1]]
- [[_COMMUNITY_Community 2|Community 2]]
- [[_COMMUNITY_Community 3|Community 3]]
- [[_COMMUNITY_Community 4|Community 4]]
- [[_COMMUNITY_Community 5|Community 5]]
- [[_COMMUNITY_Community 6|Community 6]]
- [[_COMMUNITY_Community 7|Community 7]]
- [[_COMMUNITY_Community 8|Community 8]]
- [[_COMMUNITY_Community 9|Community 9]]
- [[_COMMUNITY_Community 10|Community 10]]
- [[_COMMUNITY_Community 11|Community 11]]
- [[_COMMUNITY_Community 12|Community 12]]
- [[_COMMUNITY_Community 13|Community 13]]
- [[_COMMUNITY_Community 14|Community 14]]
- [[_COMMUNITY_Community 15|Community 15]]
- [[_COMMUNITY_Community 16|Community 16]]
- [[_COMMUNITY_Community 17|Community 17]]
- [[_COMMUNITY_Community 18|Community 18]]
- [[_COMMUNITY_Community 19|Community 19]]
- [[_COMMUNITY_Community 20|Community 20]]
- [[_COMMUNITY_Community 21|Community 21]]
- [[_COMMUNITY_Community 22|Community 22]]
- [[_COMMUNITY_Community 23|Community 23]]
- [[_COMMUNITY_Community 24|Community 24]]
- [[_COMMUNITY_Community 25|Community 25]]
- [[_COMMUNITY_Community 26|Community 26]]
- [[_COMMUNITY_Community 27|Community 27]]
- [[_COMMUNITY_Community 28|Community 28]]
- [[_COMMUNITY_Community 29|Community 29]]
- [[_COMMUNITY_Community 30|Community 30]]
- [[_COMMUNITY_Community 31|Community 31]]
- [[_COMMUNITY_Community 32|Community 32]]
- [[_COMMUNITY_Community 33|Community 33]]
- [[_COMMUNITY_Community 34|Community 34]]
- [[_COMMUNITY_Community 35|Community 35]]
- [[_COMMUNITY_Community 36|Community 36]]
- [[_COMMUNITY_Community 37|Community 37]]
- [[_COMMUNITY_Community 38|Community 38]]
- [[_COMMUNITY_Community 39|Community 39]]
- [[_COMMUNITY_Community 40|Community 40]]
- [[_COMMUNITY_Community 41|Community 41]]
- [[_COMMUNITY_Community 42|Community 42]]
- [[_COMMUNITY_Community 43|Community 43]]
- [[_COMMUNITY_Community 44|Community 44]]
- [[_COMMUNITY_Community 45|Community 45]]
- [[_COMMUNITY_Community 46|Community 46]]
- [[_COMMUNITY_Community 47|Community 47]]
- [[_COMMUNITY_Community 48|Community 48]]
- [[_COMMUNITY_Community 49|Community 49]]
- [[_COMMUNITY_Community 50|Community 50]]
- [[_COMMUNITY_Community 51|Community 51]]
- [[_COMMUNITY_Community 52|Community 52]]
- [[_COMMUNITY_Community 53|Community 53]]
- [[_COMMUNITY_Community 54|Community 54]]
- [[_COMMUNITY_Community 55|Community 55]]
- [[_COMMUNITY_Community 56|Community 56]]
- [[_COMMUNITY_Community 57|Community 57]]
- [[_COMMUNITY_Community 58|Community 58]]
- [[_COMMUNITY_Community 59|Community 59]]
- [[_COMMUNITY_Community 60|Community 60]]
- [[_COMMUNITY_Community 61|Community 61]]
- [[_COMMUNITY_Community 62|Community 62]]
- [[_COMMUNITY_Community 63|Community 63]]
- [[_COMMUNITY_Community 64|Community 64]]
- [[_COMMUNITY_Community 65|Community 65]]
- [[_COMMUNITY_Community 66|Community 66]]
- [[_COMMUNITY_Community 67|Community 67]]
- [[_COMMUNITY_Community 68|Community 68]]
- [[_COMMUNITY_Community 69|Community 69]]
- [[_COMMUNITY_Community 70|Community 70]]
- [[_COMMUNITY_Community 71|Community 71]]
- [[_COMMUNITY_Community 72|Community 72]]
- [[_COMMUNITY_Community 73|Community 73]]
- [[_COMMUNITY_Community 74|Community 74]]
- [[_COMMUNITY_Community 75|Community 75]]
- [[_COMMUNITY_Community 76|Community 76]]
- [[_COMMUNITY_Community 77|Community 77]]
- [[_COMMUNITY_Community 78|Community 78]]
- [[_COMMUNITY_Community 79|Community 79]]
- [[_COMMUNITY_Community 80|Community 80]]
- [[_COMMUNITY_Community 81|Community 81]]
- [[_COMMUNITY_Community 82|Community 82]]
- [[_COMMUNITY_Community 83|Community 83]]
- [[_COMMUNITY_Community 84|Community 84]]
- [[_COMMUNITY_Community 85|Community 85]]
- [[_COMMUNITY_Community 86|Community 86]]
- [[_COMMUNITY_Community 87|Community 87]]
- [[_COMMUNITY_Community 88|Community 88]]
- [[_COMMUNITY_Community 89|Community 89]]
- [[_COMMUNITY_Community 90|Community 90]]
- [[_COMMUNITY_Community 91|Community 91]]
- [[_COMMUNITY_Community 92|Community 92]]
- [[_COMMUNITY_Community 93|Community 93]]
- [[_COMMUNITY_Community 94|Community 94]]
- [[_COMMUNITY_Community 95|Community 95]]
- [[_COMMUNITY_Community 96|Community 96]]
- [[_COMMUNITY_Community 97|Community 97]]
- [[_COMMUNITY_Community 98|Community 98]]
- [[_COMMUNITY_Community 99|Community 99]]
- [[_COMMUNITY_Community 100|Community 100]]
- [[_COMMUNITY_Community 101|Community 101]]

## God Nodes (most connected - your core abstractions)
1. `e` - 269 edges
2. `constructor()` - 64 edges
3. `forEach()` - 43 edges
4. `L()` - 40 edges
5. `U()` - 40 edges
6. `C()` - 32 edges
7. `O()` - 32 edges
8. `has()` - 32 edges
9. `handleFetch()` - 31 edges
10. `create()` - 30 edges

## Surprising Connections (you probably didn't know these)
- `RunStream()` --calls--> `command`  [INFERRED]
  internal/logcat/logcat.go → cmd/tv/main.go
- `configCmd()` --calls--> `Load()`  [INFERRED]
  cmd/tv/main.go → internal/config/config.go
- `playerCmd()` --calls--> `Load()`  [INFERRED]
  cmd/tv/main.go → internal/config/config.go
- `playCmd()` --calls--> `Resolve()`  [INFERRED]
  cmd/tv/main.go → internal/player/detect.go
- `playCmd()` --calls--> `PlayOnTV()`  [INFERRED]
  cmd/tv/main.go → internal/player/play.go

## Import Cycles
- None detected.

## Communities (107 total, 14 thin omitted)

### Community 0 - "Community 0"
Cohesion: 0.02
Nodes (60): A_(), assertInAngularZone(), assertNotInAngularZone(), bb(), Bn(), _buildCustomControlInputCache(), Cf(), ch() (+52 more)

### Community 1 - "Community 1"
Cohesion: 0.03
Nodes (6): Db(), Dm(), e, Ib(), th(), tu()

### Community 2 - "Community 2"
Cohesion: 0.06
Nodes (62): ac(), bs(), compose(), composeAsync(), createSnapshot(), df(), dh(), Dr() (+54 more)

### Community 3 - "Community 3"
Cohesion: 0.05
Nodes (42): build, serve, test, builder, configurations, defaultConfiguration, options, cli (+34 more)

### Community 4 - "Community 4"
Cohesion: 0.08
Nodes (39): addAsyncValidators(), _allControlsDisabled(), _anyControls(), _anyControlsDirty(), _anyControlsHaveStatus(), _anyControlsTouched(), _applyFormState(), asyncValidator() (+31 more)

### Community 5 - "Community 5"
Cohesion: 0.10
Nodes (23): add32(), add32to64(), arrayBufferToWords32(), broadcast(), byteAt(), byteStringToHexString(), fk(), getLastFocusedMatchingClient() (+15 more)

### Community 6 - "Community 6"
Cohesion: 0.12
Nodes (33): attachToAppRef(), bd(), Ca(), cm(), d_(), dE(), Dt(), fd() (+25 more)

### Community 7 - "Community 7"
Cohesion: 0.06
Nodes (32): dependencies, @angular/common, @angular/compiler, @angular/core, @angular/forms, @angular/platform-browser, @angular/router, @angular/service-worker (+24 more)

### Community 8 - "Community 8"
Cohesion: 0.08
Nodes (32): ab(), ad(), Bt(), bw(), by(), cu(), ea(), ew() (+24 more)

### Community 9 - "Community 9"
Cohesion: 0.07
Nodes (28): Directory Structure, File: cmd/tv/main.go, File: docs/superpowers/plans/2026-07-27-tv-play.md, File: docs/superpowers/plans/2026-07-28-tv-monorepo-finalize.md, File: docs/superpowers/specs/2026-07-27-tv-logs-design.md, File: docs/superpowers/specs/2026-07-27-tv-remote-design.md, File: docs/WORKFLOW.md, File Format (+20 more)

### Community 10 - "Community 10"
Cohesion: 0.10
Nodes (14): append(), ke(), kill(), launch(), play(), press(), save(), sendText() (+6 more)

### Community 11 - "Community 11"
Cohesion: 0.09
Nodes (5): GC(), qC(), serialize(), Sv(), zv()

### Community 12 - "Community 12"
Cohesion: 0.08
Nodes (25): _assignAsyncValidators(), _assignValidators(), Ay(), Cb(), consumerMarkedDirty(), cw(), gw(), Ky() (+17 more)

### Community 13 - "Community 13"
Cohesion: 0.15
Nodes (22): EnsureConnected(), TestEnsureConnected_AlreadyConnected(), TestEnsureConnected_ConnectFails(), TestEnsureConnected_NotConnected(), TestEnsureConnected_Offline(), TestMockRunnerLastShellCmd(), TestMockRunnerShellWithStderrCapture(), Config (+14 more)

### Community 14 - "Community 14"
Cohesion: 0.12
Nodes (22): av(), BC(), detectContentTypeHeader(), eC(), ff(), Hb(), ia(), ir() (+14 more)

### Community 15 - "Community 15"
Cohesion: 0.11
Nodes (24): Bf(), cl(), complete(), consumerOnSignalRead(), eh(), el(), error(), Fn() (+16 more)

### Community 16 - "Community 16"
Cohesion: 0.10
Nodes (23): af(), CC(), DC(), getLogs(), gv(), Hn(), kC(), lineralizeSegments() (+15 more)

### Community 17 - "Community 17"
Cohesion: 0.13
Nodes (20): as(), au(), cg(), Co(), em(), Js(), kd(), ks() (+12 more)

### Community 18 - "Community 18"
Cohesion: 0.13
Nodes (23): cacheBust(), cacheBustedFetchFromNetwork(), constructor(), debugIdleState(), debugState(), ensureInitialized(), errorToString(), execute() (+15 more)

### Community 19 - "Community 19"
Cohesion: 0.10
Nodes (22): Ag(), applyValueToInputSignal(), Ar(), bindControlProperty(), _convertErrors(), eb(), eg(), hd() (+14 more)

### Community 20 - "Community 20"
Cohesion: 0.14
Nodes (9): assertAnError, Config, MockRunner, RealRunner, Runner, Request, ResponseWriter, Server (+1 more)

### Community 21 - "Community 21"
Cohesion: 0.13
Nodes (4): deactivateRouteAndOutlet(), Dw(), markForCheck(), wr()

### Community 22 - "Community 22"
Cohesion: 0.09
Nodes (21): assetGroups, configVersion, dataGroups, hashTable, /favicon.ico, /icons/icon-128x128.png, /icons/icon-144x144.png, /icons/icon-152x152.png (+13 more)

### Community 23 - "Community 23"
Cohesion: 0.10
Nodes (20): Angular Frontend, Architecture, Build & Run, Component tree, Embedded Static Files, Error Handling, Go Backend, New command: `cmd/tv/serve.go` (+12 more)

### Community 24 - "Community 24"
Cohesion: 0.13
Nodes (13): add(), _addParent(), aE(), _cancelExistingSubscription(), enqueue(), getBaseHref(), _hasParent(), mh() (+5 more)

### Community 25 - "Community 25"
Cohesion: 0.15
Nodes (20): create(), createComponent(), createComponentRef(), createEmbeddedView(), dg(), fu(), gd(), i_() (+12 more)

### Community 26 - "Community 26"
Cohesion: 0.22
Nodes (18): Config, Runner, T, ActivityForPlayer(), EscapeURLForShell(), MimeForURL(), PlayOnTV(), TestActivityForPlayer_Known() (+10 more)

### Community 28 - "Community 28"
Cohesion: 0.22
Nodes (18): bv(), capture(), consumeOptional(), ef(), FC(), nullValidator(), parse(), parseChildren() (+10 more)

### Community 29 - "Community 29"
Cohesion: 0.18
Nodes (19): am(), bh(), Ct(), Er(), fm(), gu(), hm(), Jt() (+11 more)

### Community 30 - "Community 30"
Cohesion: 0.13
Nodes (19): dispatchEvent(), Dp(), element(), Fb(), fg(), fw(), ig(), ip() (+11 more)

### Community 31 - "Community 31"
Cohesion: 0.11
Nodes (18): Architecture, Channel, CLI handler (cmd/tv/main.go), Command Set (Flat), Files Changed/Created, Goal, Implementation, Key Mapping (+10 more)

### Community 32 - "Community 32"
Cohesion: 0.25
Nodes (17): Config, Runner, T, BuildArgs(), ResolvePID(), RunOnce(), RunStream(), TestBuildArgs_Clear() (+9 more)

### Community 33 - "Community 33"
Cohesion: 0.37
Nodes (6): Request, ResponseWriter, Server, ResponseWriter, jsonError(), jsonOK()

### Community 34 - "Community 34"
Cohesion: 0.14
Nodes (18): ao(), attach(), ed(), Gb(), hi(), ka(), kg(), Lu() (+10 more)

### Community 35 - "Community 35"
Cohesion: 0.17
Nodes (17): En(), gE(), io(), la(), md(), Mt(), Or(), Pr() (+9 more)

### Community 36 - "Community 36"
Cohesion: 0.12
Nodes (16): Architecture Issues (5), Bugs Found (3), Edge Cases Documented (not code changes), Structure Changes, Sub-task 5a: Fix BuildArgs to support --pid filter, Sub-task 5b: Fix cmd/tv/main.go to pass resolved PID to logcat, Task 1: Move pkg/* → internal/* (structural fix), Task 2: Fix duplicate loadConfig() in adb package (+8 more)

### Community 37 - "Community 37"
Cohesion: 0.25
Nodes (13): at(), emit(), getLView(), gp(), Hu(), insert(), move(), swap() (+5 more)

### Community 38 - "Community 38"
Cohesion: 0.21
Nodes (3): path(), Sn(), wm()

### Community 39 - "Community 39"
Cohesion: 0.20
Nodes (12): appendChild(), constructor(), createElement(), createText(), getDefaultDocument(), _initObservables(), jd(), ll() (+4 more)

### Community 40 - "Community 40"
Cohesion: 0.22
Nodes (13): abortInProgressLoad(), addHeaderEntry(), ap(), applyUpdate(), has(), maybeSetNormalizedName(), mw(), pw() (+5 more)

### Community 41 - "Community 41"
Cohesion: 0.19
Nodes (15): cE(), im(), iS(), jI(), kI(), loadEffect(), Me(), nl() (+7 more)

### Community 42 - "Community 42"
Cohesion: 0.20
Nodes (15): C(), destroyed(), eu(), jn(), kp(), ln(), Po(), producerMustRecompute() (+7 more)

### Community 43 - "Community 43"
Cohesion: 0.25
Nodes (15): Bg(), cd(), Cn(), dd(), di(), hg(), lw(), oi() (+7 more)

### Community 44 - "Community 44"
Cohesion: 0.26
Nodes (15): accessed(), cacheResponse(), cacheStatus(), clearCacheForUrl(), delete(), fetchFromCacheOnly(), loadFromCache(), match() (+7 more)

### Community 45 - "Community 45"
Cohesion: 0.17
Nodes (15): Al(), BE(), Dl(), fh(), hE(), Il(), injectableDefInScope(), lE() (+7 more)

### Community 46 - "Community 46"
Cohesion: 0.15
Nodes (15): bl(), cp(), cs(), Fs(), gl(), Gn(), $h(), lh() (+7 more)

### Community 47 - "Community 47"
Cohesion: 0.28
Nodes (13): Config, Runner, T, Detect(), Resolve(), TestDetect_Known(), TestDetect_PrefixMatch(), TestDetect_Sorted() (+5 more)

### Community 48 - "Community 48"
Cohesion: 0.20
Nodes (13): _adjustIndex(), attachToViewContainerRef(), bu(), detach(), detachFromAppRef(), gi(), II(), insertImpl() (+5 more)

### Community 49 - "Community 49"
Cohesion: 0.25
Nodes (3): clear(), destroy(), qm()

### Community 50 - "Community 50"
Cohesion: 0.20
Nodes (5): decoratePreventDefault(), getGlobalEventTarget(), JE(), listen(), onAndCancel()

### Community 51 - "Community 51"
Cohesion: 0.24
Nodes (11): GetTVIP(), Config, Dir(), Load(), path(), TestDir(), TestLoad_ReturnsDefaultsWhenMissing(), TestSaveAndLoad() (+3 more)

### Community 52 - "Community 52"
Cohesion: 0.21
Nodes (11): activate(), activateChildRoutes(), activateRoutes(), deactivateChildRoutes(), deactivateRouteAndItsChildren(), deactivateRoutes(), detachAndStoreRouteSubtree(), nf() (+3 more)

### Community 53 - "Community 53"
Cohesion: 0.22
Nodes (11): __(), appendAll(), clone(), copyFrom(), delete(), forEach(), getAll(), init() (+3 more)

### Community 54 - "Community 54"
Cohesion: 0.18
Nodes (3): LogsPage, LogFrame, WebSocketService

### Community 55 - "Community 55"
Cohesion: 0.24
Nodes (10): Context, FS, Handler, Config, Runner, Server, corsMiddleware(), New() (+2 more)

### Community 56 - "Community 56"
Cohesion: 0.18
Nodes (12): ah(), aI(), bm(), Br(), ci(), da(), fE(), getCookie() (+4 more)

### Community 57 - "Community 57"
Cohesion: 0.18
Nodes (11): addClass(), aw(), indexOf(), ow(), removeClass(), removeOnDestroy(), removeStyle(), setStyle() (+3 more)

### Community 58 - "Community 58"
Cohesion: 0.30
Nodes (11): T, TestKeys_ReturnsSorted(), TestPress_MediaKey(), TestPress_NavKey(), TestPress_NumberKey(), TestPress_SystemKey(), TestPress_UnknownKey(), TestPress_VolumeKey() (+3 more)

### Community 59 - "Community 59"
Cohesion: 0.17
Nodes (11): Architecture, Argument Building, Command Set, Files, Goal, Implementation, PID Resolution, pkg/logcat/logcat.go (+3 more)

### Community 60 - "Community 60"
Cohesion: 0.20
Nodes (9): getChildConfig(), hC(), iu(), p(), pT(), retrieve(), _setUpdateStrategy(), vM() (+1 more)

### Community 61 - "Community 61"
Cohesion: 0.20
Nodes (11): Fv(), Hv(), removeAsyncValidators(), removeParseErrorsValidator(), removeValidators(), tf(), toRoot(), Uv() (+3 more)

### Community 62 - "Community 62"
Cohesion: 0.18
Nodes (10): Task 1: Go internal/serve/server.go — HTTP server with embedded static files, Task 2: REST API handlers, Task 3: WebSocket handler for logcat, Task 4: tv serve CLI command, Task 5: Scaffold Angular PWA project (following angular-new-app skill), Task 6: Angular services (scaffolded in Task 5, now implement), Task 7: Angular UI pages (scaffolded in Task 5, now implement), Task 8: Build pipeline and integration (+2 more)

### Community 63 - "Community 63"
Cohesion: 0.22
Nodes (3): flush(), flushQueue(), ni()

### Community 64 - "Community 64"
Cohesion: 0.24
Nodes (5): addValidators(), ey(), Ho(), hw(), setupCustomControl()

### Community 65 - "Community 65"
Cohesion: 0.24
Nodes (10): acceptsTextHtml(), assignVersion(), has(), isNavigationRequest(), lookupResourceWithHash(), lookupResourceWithoutHash(), lookupVersionByHash(), maybeUpdate() (+2 more)

### Community 66 - "Community 66"
Cohesion: 0.31
Nodes (10): detectStorageFull(), handleFetchWithFreshness(), handleFetchWithPerformance(), log(), networkFetchWithTimeout(), newResponse(), onMessageError(), onUnhandledRejection() (+2 more)

### Community 67 - "Community 67"
Cohesion: 0.22
Nodes (10): checkForUpdate(), deleteAllCaches(), fetchLatestManifest(), initializeFully(), notifyClientsAboutNoNewVersionDetected(), notifyClientsAboutVersionDetected(), notifyClientsAboutVersionInstallationFailed(), notifyClientsAboutVersionReady() (+2 more)

### Community 68 - "Community 68"
Cohesion: 0.22
Nodes (10): bi(), createComment(), Dn(), Ei(), insertBefore(), Km(), nextSibling(), nodeOrShadowRoot() (+2 more)

### Community 70 - "Community 70"
Cohesion: 0.33
Nodes (9): Aa(), fI(), IE(), ml(), oo(), qp(), uI(), Ur() (+1 more)

### Community 71 - "Community 71"
Cohesion: 0.36
Nodes (9): _b(), Bo(), bp(), get(), gt(), pE(), registerOnDisabledChange(), resolveInjectorInitializers() (+1 more)

### Community 72 - "Community 72"
Cohesion: 0.22
Nodes (9): cleanupCaches(), completeOperation(), getCacheNames(), handleMessage(), isMsgActivateUpdate(), isMsgCheckForUpdates(), mergeHashWithAppData(), sync() (+1 more)

### Community 73 - "Community 73"
Cohesion: 0.22
Nodes (8): Task 1: Module init + directory structure, Task 2: pkg/config — shared config package, Task 3: pkg/adb — shared ADB interface + mock, Task 4: pkg/player — detection + playback logic, Task 5: cmd/tv — parent CLI entry point, Task 6: Build artifacts, Task 7: Integration + GitHub push, tv Implementation Plan

### Community 75 - "Community 75"
Cohesion: 0.29
Nodes (8): fl(), hl(), lg(), Mo(), ol(), run(), ug(), Zo()

### Community 76 - "Community 76"
Cohesion: 0.29
Nodes (8): debugVersions(), initialize(), isLocalhost(), keys(), list(), schedule(), scheduleInitialization(), unhashedResources()

### Community 77 - "Community 77"
Cohesion: 0.39
Nodes (3): App, appConfig, routes

### Community 78 - "Community 78"
Cohesion: 0.25
Nodes (7): Accessibility Requirements, Angular Best Practices, Components, Services, State Management, Templates, TypeScript Best Practices

### Community 79 - "Community 79"
Cohesion: 0.25
Nodes (7): Additional Resources, Building, Code scaffolding, Development server, Running end-to-end tests, Running unit tests, Web

### Community 82 - "Community 82"
Cohesion: 0.33
Nodes (4): cy(), encodeKey(), encodeValue(), nv()

### Community 83 - "Community 83"
Cohesion: 0.47
Nodes (6): du(), li(), op(), VI(), Xo(), yI()

### Community 84 - "Community 84"
Cohesion: 0.40
Nodes (4): generateNonce(), postMessage(), postMessageWithOperation(), waitForOperationCompleted()

### Community 85 - "Community 85"
Cohesion: 0.53
Nodes (5): Config, Runner, Keys(), Press(), Type()

### Community 86 - "Community 86"
Cohesion: 0.33
Nodes (5): configRequest, packageRequest, playRequest, pressRequest, typeRequest

### Community 87 - "Community 87"
Cohesion: 0.33
Nodes (4): AppInfo, LogResult, Status, TvConfig

### Community 88 - "Community 88"
Cohesion: 0.40
Nodes (4): Branch, Git Worktrees, GitHub, Workflow Conventions

### Community 89 - "Community 89"
Cohesion: 0.40
Nodes (5): applyRedirectCommands(), applyRedirectCreateUrlTree(), createQueryParams(), createSegmentGroup(), createSegments()

### Community 90 - "Community 90"
Cohesion: 0.40
Nodes (5): hasAsyncValidator(), hasValidator(), Jc(), tE(), Vf()

### Community 91 - "Community 91"
Cohesion: 0.40
Nodes (5): connect(), filteredApps(), getValue(), ngOnInit(), value()

### Community 95 - "Community 95"
Cohesion: 0.50
Nodes (3): assetGroups, index, $schema

## Knowledge Gaps
- **219 isolated node(s):** `Config`, `Config`, `Runner`, `pressRequest`, `typeRequest` (+214 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **14 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `e` connect `Community 1` to `Community 0`, `Community 2`, `Community 4`, `Community 6`, `Community 8`, `Community 10`, `Community 11`, `Community 14`, `Community 16`, `Community 17`, `Community 19`, `Community 21`, `Community 24`, `Community 25`, `Community 27`, `Community 28`, `Community 29`, `Community 34`, `Community 35`, `Community 37`, `Community 38`, `Community 39`, `Community 40`, `Community 41`, `Community 42`, `Community 45`, `Community 48`, `Community 49`, `Community 50`, `Community 52`, `Community 53`, `Community 57`, `Community 60`, `Community 63`, `Community 64`, `Community 74`, `Community 81`, `Community 82`, `Community 84`?**
  _High betweenness centrality (0.120) - this node is a cross-community bridge._
- **Why does `destroy()` connect `Community 49` to `Community 0`, `Community 1`, `Community 71`, `Community 40`, `Community 8`, `Community 42`, `Community 43`, `Community 15`, `Community 48`, `Community 52`, `Community 21`, `Community 53`, `Community 57`?**
  _High betweenness centrality (0.003) - this node is a cross-community bridge._
- **Why does `forEach()` connect `Community 53` to `Community 0`, `Community 2`, `Community 4`, `Community 8`, `Community 14`, `Community 15`, `Community 17`, `Community 24`, `Community 28`, `Community 29`, `Community 37`, `Community 38`, `Community 39`, `Community 40`, `Community 41`, `Community 42`, `Community 49`, `Community 50`, `Community 52`, `Community 57`, `Community 70`, `Community 81`, `Community 83`, `Community 89`?**
  _High betweenness centrality (0.002) - this node is a cross-community bridge._
- **Are the 12 inferred relationships involving `e` (e.g. with `cE()` and `Dm()`) actually correct?**
  _`e` has 12 INFERRED edges - model-reasoned connections that need verification._
- **Are the 20 inferred relationships involving `constructor()` (e.g. with `Bf()` and `cl()`) actually correct?**
  _`constructor()` has 20 INFERRED edges - model-reasoned connections that need verification._
- **Are the 9 inferred relationships involving `L()` (e.g. with `cg()` and `kg()`) actually correct?**
  _`L()` has 9 INFERRED edges - model-reasoned connections that need verification._
- **Are the 3 inferred relationships involving `U()` (e.g. with `cd()` and `ha()`) actually correct?**
  _`U()` has 3 INFERRED edges - model-reasoned connections that need verification._