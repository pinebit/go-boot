# go-boot

A micro-framework for deterministic and graceful services boot and shutdown.

# Why?

Most golang applications are made of multiple services, where each service is an entity that can be started and stopped on-demand. Most of the services depend on each other, for example it makes little sense to start HTTP server until a DB connection is established. Therefore, all the services need to be started in the right order and then stopped in the reverse order. This allows application to boot and shutdown gracefully and deterministically, properly acquiring and releasing all the resources, locks, etc.

Furthermore, each application has to respect OS signals (e.g. SIGINT, SIGTERM). And such the signals can be caught even during the application boot process and in that case they should trigger the shutdown flow immediately. The shutdown flow is also non-trivial, because for most execution environments an application is given a little window to shutdown gracefully (usually 5 to 15 seconds) before it gets SIGKILL. All the flows are usually controlled by a `context` provided by the application, either bound to OS signals, or a timeout. 

This architecure pattern became very common for most applications, and this is why `go-boot` was built as a standalone micro-framework.

# Usage

1. Installation

```shell
go get -u github.com/pinebit/go-boot
```

2. Make your services conforming the `boot.Service` interface

```golang
package service1

type Service1 interface {
    boot.Service
}
```

3. Instantiate all your services

> Recommended to use Dependency Injection frameworks, such as [wire](https://github.com/google/wire), [fx](https://github.com/uber-go/fx) or any other, for now we simply create all services one by one.

```golang
s1 := service1.NewService1(...)
s2 := service2.NewService2(...)
s3 := service3.NewService3(...) 
```

4. Start all services respecting the given `context`

```golang
// You can choose between boot.Sequentially, boot.Simultaneously or combine them.
services := boot.Sequentially(s1, s2, s3)
// the ctx can stop the boot flow gracefully
if err := services.Start(ctx); err != nil {
    // report error and proceed with the shutdown 
}
```

5. Shutdown all services respecting a timeout

```golang
// The recommended shutdown timeout is five seconds for most systems.
ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
defer cancel()
// Stop() will stop all services in the reverse order.
// The shutdown flow would break if the ctx is done.
if err := services.Stop(ctx); err != nil {
    // One of the services reported error, or ctx is done.
    // Log the error for further investigation.
    // You still have a good chance to release critical resources, locks. 
}
```

# Advanced usage

You can combine sequential and parallel boot flows. For example, given services A, B, C and D. Where service A must be started first, then B and C can be started simultaneously and then D can be started only after A, B and C have started:

```golang
services := boot.Sequentially(a, boot.Simultaneously(b, c), d)
err := services.Start(ctx)
```

# License

MIT. See the `LICENSE` file for details.
