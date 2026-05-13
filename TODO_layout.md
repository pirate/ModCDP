repo/
    README.md
    docs/
    bin/

    extension/
      manifest.json
      dist/
        extension/
        extension.zip
      src/
        config.ts
        service_worker.ts   <- this should be nearly empty, it just sets up ModCDPServer.ts and then delegates to it
        // ModCDPServer.ts  <- this belongs in js/src/server/ModCDPServer.ts not in the extension, because other projects may define their own extensions that need to extend ModCDPServer.ts

        pages/
          options.html
          options.ts
          wake.html
          wake.ts
          offscreen_keepalive.html
          offscreen_keepalive.ts

    js/
      package.json
      tsconfig.json
      dist/
        modcdp/...
      src/
        index.ts

        server/
          ModCDPServer.ts

        client/
          ModCDPClient.ts

        launcher/
          BrowserLauncher.ts
          LocalBrowserLauncher.ts
          RemoteBrowserLauncher.ts
          BrowserbaseBrowserLauncher.ts
          NoopBrowserLauncher.ts

        injector/
          ExtensionInjector.ts
          LocalBrowserLaunchExtensionInjector.ts
          ExtensionsLoadUnpackedInjector.ts
          DiscoveredExtensionInjector.ts
          BorrowedExtensionInjector.ts
          BBBrowserExtensionInjector.ts

        transport/
          UpstreamTransport.ts
          WebSocketUpstreamTransport.ts
          ReverseWebSocketUpstreamTransport.ts
          NativeMessagingUpstreamTransport.ts
          NatsUpstreamTransport.ts
          PipeUpstreamTransport.ts

        router/
          AutoSessionRouter.ts

        translate/
          translate.ts

        types/
          modcdp.ts
          codegen.ts
          generated/
            aliases.ts
            cdp.ts
            zod.ts
            zod/
              Accessibility.ts
              Animation.ts
              ...

        proxy/
          proxy.ts
          cli.ts

      examples/
        demo.ts
        playwright.ts
        puppeteer.ts

      test/
        test.ModCDPClient.ts
        test.LocalBrowserLauncher.ts
        test.WebSocketUpstreamTransport.ts
        test.translate.ts

    python/
      pyproject.toml
      README.md
      dist/...
      
      modcdp/
        __init__.py

        client/
          __init__.py
          ModCDPClient.py

        launcher/
          __init__.py
          BrowserLauncher.py
          LocalBrowserLauncher.py
          RemoteBrowserLauncher.py
          BrowserbaseBrowserLauncher.py
          NoopBrowserLauncher.py

        injector/
          __init__.py
          ExtensionInjector.py
          LocalBrowserLaunchExtensionInjector.py
          ExtensionsLoadUnpackedInjector.py
          DiscoveredExtensionInjector.py
          BorrowedExtensionInjector.py
          BBBrowserExtensionInjector.py

        transport/
          __init__.py
          UpstreamTransport.py
          WebSocketUpstreamTransport.py
          ReverseWebSocketUpstreamTransport.py
          NativeMessagingUpstreamTransport.py
          NatsUpstreamTransport.py
          PipeUpstreamTransport.py

        router/
          __init__.py
          AutoSessionRouter.py

        translate/
          __init__.py
          translate.py

        types/
          __init__.py
          codegen.py
          modcdp.py
          jsonschema.py
          generated/
            __init__.py
            cdp.py

      examples/
        demo.py

      tests/
        test_ModCDPClient.py
        test_LocalBrowserLauncher.py
        test_WebSocketUpstreamTransport.py
        test_translate.py
        ...
        test_{ClassName}.py

    go/
      go.mod
      go.sum
      modcdp/
        modcdp.go

        client/
          ModCDPClient.go
          ModCDPClient_test.go
          // generated CDP surface stays in client package because it defines methods on ModCDPClient
          generated.go
          generated_domains.go

        launcher/
          BrowserLauncher.go
          BrowserLauncher_test.go
          LocalBrowserLauncher.go
          LocalBrowserLauncher_test.go
          RemoteBrowserLauncher.go
          RemoteBrowserLauncher_test.go
          BrowserbaseBrowserLauncher.go
          BrowserbaseBrowserLauncher_test.go
          NoopBrowserLauncher.go
          NoopBrowserLauncher_test.go

        injector/
          ExtensionInjector.go
          ExtensionInjector_test.go
          LocalBrowserLaunchExtensionInjector.go
          LocalBrowserLaunchExtensionInjector_test.go
          ExtensionsLoadUnpackedInjector.go
          ExtensionsLoadUnpackedInjector_test.go
          DiscoveredExtensionInjector.go
          DiscoveredExtensionInjector_test.go
          BorrowedExtensionInjector.go
          BorrowedExtensionInjector_test.go
          BBBrowserExtensionInjector.go
          BBBrowserExtensionInjector_test.go

        transport/
          UpstreamTransport.go
          UpstreamTransport_test.go
          WebSocketUpstreamTransport.go
          WebSocketUpstreamTransport_test.go
          ReverseWebSocketUpstreamTransport.go
          ReverseWebSocketUpstreamTransport_test.go
          NativeMessagingUpstreamTransport.go
          NativeMessagingUpstreamTransport_test.go
          NatsUpstreamTransport.go
          NatsUpstreamTransport_test.go
          PipeUpstreamTransport.go
          PipeUpstreamTransport_test.go

        router/
          AutoSessionRouter.go
          AutoSessionRouter_test.go

        translate/
          translate.go
          translate_test.go

        types/
          codegen.go
          types.go

      examples/
        demo/
          main.go
