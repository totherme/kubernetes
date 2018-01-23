load("@io_bazel_rules_go//go:def.bzl", "go_test")

go_test(
    name = "go_default_xtest",
    srcs = [
        "kubectl_integration_test.go",
        "kubectl_suite_test.go",
    ],
    deps = [
        "//vendor/github.com/kubernetes-sig-testing/frameworks/integration:go_default_library",
        "//vendor/github.com/onsi/ginkgo:go_default_library",
        "//vendor/github.com/onsi/gomega:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/util/sets:go_default_library",
    ],
)

filegroup(
    name = "package-srcs",
    srcs = glob(["**"]),
    tags = ["automanaged"],
    visibility = ["//visibility:private"],
)

filegroup(
    name = "all-srcs",
    srcs = [":package-srcs"],
    tags = ["automanaged"],
    visibility = ["//visibility:public"],
)
