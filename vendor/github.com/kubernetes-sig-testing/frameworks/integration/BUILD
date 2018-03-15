load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "apiserver.go",
        "control_plane.go",
        "doc.go",
        "etcd.go",
        "kubectl.go",
    ],
    importpath = "github.com/kubernetes-sig-testing/frameworks/integration",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/kubernetes-sig-testing/frameworks/integration/internal:go_default_library"],
)

filegroup(
    name = "package-srcs",
    srcs = glob(["**"]),
    tags = ["automanaged"],
    visibility = ["//visibility:private"],
)

filegroup(
    name = "all-srcs",
    srcs = [
        ":package-srcs",
        "//vendor/github.com/kubernetes-sig-testing/frameworks/integration/internal:all-srcs",
    ],
    tags = ["automanaged"],
    visibility = ["//visibility:public"],
)
