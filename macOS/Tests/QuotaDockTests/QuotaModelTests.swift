import XCTest
@testable import QuotaDock

final class QuotaModelTests: XCTestCase {
    func testPercentageCalculation() {
        let telemetry = ProviderTelemetry(
            provider: "anthropic",
            limitRequests: 1000,
            remainingRequests: 500,
            limitTokens: 1000000,
            remainingTokens: 250000,
            tokensPerSecond: 100,
            estimatedSecondsRemaining: 2500,
            status: "warning"
        )

        XCTAssertEqual(telemetry.percentageRemaining, 0.25, accuracy: 0.001)
    }

    func testZeroLimitHandling() {
        let telemetry = ProviderTelemetry(
            provider: "local",
            limitRequests: 0,
            remainingRequests: 0,
            limitTokens: 0,
            remainingTokens: 0,
            tokensPerSecond: 0,
            estimatedSecondsRemaining: 0,
            status: "healthy"
        )

        XCTAssertEqual(telemetry.percentageRemaining, 1.0)
    }
}
