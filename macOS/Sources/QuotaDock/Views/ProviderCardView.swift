import SwiftUI

public struct ProviderCardView: View {
    public let item: ProviderTelemetry
    @State private var isHovered: Bool = false

    public init(item: ProviderTelemetry) {
        self.item = item
    }

    private var providerIcon: String {
        switch item.provider.lowercased() {
        case "anthropic": return "brain.head.profile"
        case "openai": return "sparkles"
        case "gemini": return "diamond.fill"
        case "ollama": return "server.rack"
        default: return "cpu"
        }
    }

    private var formattedTokens: String {
        let count = Double(item.remainingTokens)
        if count >= 1_000_000 {
            return String(format: "%.1fM", count / 1_000_000)
        } else if count >= 1_000 {
            return String(format: "%.0fK", count / 1_000)
        }
        return "\(item.remainingTokens)"
    }

    private var formattedMaxTokens: String {
        let count = Double(item.limitTokens)
        if count >= 1_000_000 {
            return String(format: "%.1fM", count / 1_000_000)
        } else if count >= 1_000 {
            return String(format: "%.0fK", count / 1_000)
        }
        return "\(item.limitTokens)"
    }

    public var body: some View {
        HStack(spacing: 12) {
            // Circular Ring Gauge
            ZStack {
                Circle()
                    .stroke(Color.primary.opacity(0.08), lineWidth: 3.5)
                Circle()
                    .trim(from: 0, to: CGFloat(min(max(item.percentageRemaining, 0.0), 1.0)))
                    .stroke(
                        item.statusColor,
                        style: StrokeStyle(lineWidth: 3.5, lineCap: .round)
                    )
                    .rotationEffect(.degrees(-90))
                    .animation(.spring(response: 0.35, dampingFraction: 0.75), value: item.percentageRemaining)

                Image(systemName: providerIcon)
                    .font(.system(size: 11, weight: .medium))
                    .foregroundColor(.secondary)
            }
            .frame(width: 32, height: 32)

            // Info Column
            VStack(alignment: .leading, spacing: 2) {
                HStack(alignment: .firstTextBaseline) {
                    Text(item.provider.capitalized)
                        .font(.system(size: 13, weight: .semibold, design: .default))
                        .foregroundColor(.primary)

                    Spacer()

                    // Traffic Light Badge
                    Circle()
                        .fill(item.statusColor)
                        .frame(width: 6, height: 6)

                    Text(item.status.capitalized)
                        .font(.system(size: 10, weight: .medium))
                        .foregroundColor(item.statusColor)
                }

                // Tokens Counter with tabular figures
                HStack(spacing: 4) {
                    Text("\(formattedTokens) / \(formattedMaxTokens) tokens")
                        .font(.system(size: 11, weight: .regular, design: .monospaced))
                        .monospacedDigit()
                        .foregroundColor(.secondary)

                    Spacer()

                    if item.tokensPerSecond > 0 {
                        Text("\(Int(item.tokensPerSecond)) t/s")
                            .font(.system(size: 10, weight: .medium, design: .monospaced))
                            .monospacedDigit()
                            .foregroundColor(.secondary.opacity(0.8))
                    }
                }
            }
        }
        .padding(10)
        .background(
            RoundedRectangle(cornerRadius: 10, style: .continuous)
                .fill(isHovered ? Color.primary.opacity(0.06) : Color.primary.opacity(0.03))
        )
        .overlay(
            RoundedRectangle(cornerRadius: 10, style: .continuous)
                .stroke(Color.primary.opacity(0.08), lineWidth: 0.5)
        )
        .onHover { hovering in
            withAnimation(.easeInOut(duration: 0.15)) {
                self.isHovered = hovering
            }
        }
    }
}
