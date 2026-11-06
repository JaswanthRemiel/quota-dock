import SwiftUI
import AppKit

public struct MenuBarView: View {
    @ObservedObject var store: QuotaStore
    @State private var showingSettings: Bool = false

    public init(store: QuotaStore) {
        self.store = store
    }

    public var body: some View {
        VStack(spacing: 12) {
            // Header
            HStack(alignment: .center) {
                HStack(spacing: 6) {
                    Circle()
                        .fill(store.isDaemonConnected ? Color(red: 0.20, green: 0.78, blue: 0.35) : Color(red: 0.95, green: 0.26, blue: 0.21))
                        .frame(width: 8, height: 8)

                    Text("QuotaDock")
                        .font(.system(size: 13, weight: .bold))
                }

                Spacer()

                Button {
                    showingSettings.toggle()
                } label: {
                    Image(systemName: "gearshape")
                        .font(.system(size: 12, weight: .medium))
                        .foregroundColor(.secondary)
                }
                .buttonStyle(.plain)
                .sheet(isPresented: $showingSettings) {
                    SettingsView()
                }

                Button {
                    NSApplication.shared.terminate(nil)
                } label: {
                    Image(systemName: "power")
                        .font(.system(size: 12, weight: .medium))
                        .foregroundColor(.secondary)
                }
                .buttonStyle(.plain)
            }
            .padding(.horizontal, 4)

            // Provider Cards
            VStack(spacing: 8) {
                ForEach(store.quotas) { quota in
                    ProviderCardView(item: quota)
                }
            }

            // Divider line
            Rectangle()
                .fill(Color.primary.opacity(0.08))
                .frame(height: 1)
                .padding(.horizontal, 2)

            // Failover Chain Status Footer
            VStack(alignment: .leading, spacing: 6) {
                HStack {
                    Text("ACTIVE FAILOVER CHAIN")
                        .font(.system(size: 9, weight: .semibold))
                        .foregroundColor(.secondary)

                    Spacer()

                    Text("localhost:8080")
                        .font(.system(size: 9, design: .monospaced))
                        .foregroundColor(.secondary.opacity(0.8))
                }

                HStack(spacing: 4) {
                    chainPill(name: "Claude", active: true)
                    Image(systemName: "arrow.right")
                        .font(.system(size: 9))
                        .foregroundColor(.secondary)
                    chainPill(name: "GPT-4o", active: false)
                    Image(systemName: "arrow.right")
                        .font(.system(size: 9))
                        .foregroundColor(.secondary)
                    chainPill(name: "Gemini", active: false)
                    Image(systemName: "arrow.right")
                        .font(.system(size: 9))
                        .foregroundColor(.secondary)
                    chainPill(name: "Ollama", active: false)
                }
            }
            .padding(10)
            .background(
                RoundedRectangle(cornerRadius: 8, style: .continuous)
                    .fill(Color.primary.opacity(0.02))
            )
        }
        .padding(14)
        .frame(width: 340)
        .background(Material.ultraThinMaterial)
    }

    private func chainPill(name: String, active: Bool) -> some View {
        Text(name)
            .font(.system(size: 10, weight: active ? .semibold : .regular))
            .foregroundColor(active ? .primary : .secondary)
            .padding(.horizontal, 6)
            .padding(.vertical, 3)
            .background(
                Capsule()
                    .fill(active ? Color.accentColor.opacity(0.15) : Color.primary.opacity(0.04))
            )
    }
}
