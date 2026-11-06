// swift-tools-version: 6.0
import PackageDescription

let package = Package(
    name: "QuotaDock",
    platforms: [
        .macOS(.v13)
    ],
    products: [
        .executable(name: "QuotaDock", targets: ["QuotaDock"])
    ],
    targets: [
        .executableTarget(
            name: "QuotaDock",
            path: "Sources/QuotaDock"
        ),
        .testTarget(
            name: "QuotaDockTests",
            dependencies: ["QuotaDock"],
            path: "Tests/QuotaDockTests"
        )
    ]
)
