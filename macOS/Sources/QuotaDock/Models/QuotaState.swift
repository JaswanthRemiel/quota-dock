import Foundation
import SwiftUI
import Combine

public struct ProviderTelemetry: Identifiable, Codable, Equatable {
    public var id: String { provider }
    public let provider: String
    public let limitRequests: Int64
    public let remainingRequests: Int64
    public let limitTokens: Int64
    public let remainingTokens: Int64
    public let tokensPerSecond: Double
    public let estimatedSecondsRemaining: Double
    public let status: String // "healthy", "warning", "critical", "exhausted"

    enum CodingKeys: String, CodingKey {
        case provider
        case limitRequests = "limit_requests"
        case remainingRequests = "remaining_requests"
        case limitTokens = "limit_tokens"
        case remainingTokens = "remaining_tokens"
        case tokensPerSecond = "tokens_per_second"
        case estimatedSecondsRemaining = "estimated_seconds_remaining"
        case status
    }

    public var percentageRemaining: Double {
        guard limitTokens > 0 else { return 1.0 }
        return Double(remainingTokens) / Double(limitTokens)
    }

    public var statusColor: Color {
        switch status.lowercased() {
        case "healthy":
            return Color(red: 0.20, green: 0.78, blue: 0.35) // Emerald Green
        case "warning":
            return Color(red: 1.0, green: 0.65, blue: 0.0)   // Amber
        case "critical", "exhausted":
            return Color(red: 0.95, green: 0.26, blue: 0.21) // Crimson Red
        default:
            return Color.secondary.opacity(0.4)
        }
    }
}

public struct TelemetryResponse: Codable {
    public let quotas: [ProviderTelemetry]
}

@MainActor
public final class QuotaStore: ObservableObject {
    @Published public var quotas: [ProviderTelemetry] = []
    @Published public var isDaemonConnected: Bool = false
    @Published public var overallHealthStatus: String = "healthy"
    
    private var timer: AnyCancellable?
    private let telemetryEndpoint = URL(string: "http://127.0.0.1:8080/quotadock/telemetry")!

    public init() {
        setupDemoDefaults()
        startPolling()
    }

    private func setupDemoDefaults() {
        self.quotas = [
            ProviderTelemetry(
                provider: "anthropic",
                limitRequests: 1000,
                remainingRequests: 840,
                limitTokens: 1000000,
                remainingTokens: 780000,
                tokensPerSecond: 145.2,
                estimatedSecondsRemaining: 5371,
                status: "healthy"
            ),
            ProviderTelemetry(
                provider: "openai",
                limitRequests: 500,
                remainingRequests: 420,
                limitTokens: 2000000,
                remainingTokens: 1450000,
                tokensPerSecond: 64.0,
                estimatedSecondsRemaining: 22656,
                status: "healthy"
            ),
            ProviderTelemetry(
                provider: "gemini",
                limitRequests: 1500,
                remainingRequests: 1200,
                limitTokens: 4000000,
                remainingTokens: 3890000,
                tokensPerSecond: 0.0,
                estimatedSecondsRemaining: 0.0,
                status: "healthy"
            ),
            ProviderTelemetry(
                provider: "ollama",
                limitRequests: 999999,
                remainingRequests: 999999,
                limitTokens: 99999999,
                remainingTokens: 99999999,
                tokensPerSecond: 0.0,
                estimatedSecondsRemaining: 0.0,
                status: "healthy"
            )
        ]
    }

    public func startPolling() {
        timer = Timer.publish(every: 1.5, on: .main, in: .common)
            .autoconnect()
            .sink { [weak self] _ in
                Task { [weak self] in
                    await self?.fetchTelemetry()
                }
            }
    }

    public func fetchTelemetry() async {
        do {
            let (data, response) = try await URLSession.shared.data(from: telemetryEndpoint)
            guard let httpResp = response as? HTTPURLResponse, httpResp.statusCode == 200 else {
                self.isDaemonConnected = false
                return
            }
            let decoded = try JSONDecoder().decode(TelemetryResponse.self, from: data)
            self.quotas = decoded.quotas
            self.isDaemonConnected = true
            updateOverallHealth()
        } catch {
            self.isDaemonConnected = false
        }
    }

    private func updateOverallHealth() {
        if quotas.contains(where: { $0.status == "exhausted" || $0.status == "critical" }) {
            self.overallHealthStatus = "warning"
        } else {
            self.overallHealthStatus = "healthy"
        }
    }
}
