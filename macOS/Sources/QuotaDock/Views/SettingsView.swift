import SwiftUI

public struct SettingsView: View {
    @Environment(\.dismiss) private var dismiss
    @State private var anthropicKey: String = ""
    @State private var openaiKey: String = ""
    @State private var geminiKey: String = ""
    @State private var saveMessage: String?

    public init() {}

    public var body: some View {
        VStack(alignment: .leading, spacing: 16) {
            // Header
            HStack {
                Text("API Credentials (macOS Keychain)")
                    .font(.system(size: 13, weight: .semibold))
                Spacer()
                Button("Done") {
                    dismiss()
                }
                .buttonStyle(.borderless)
                .font(.system(size: 12, weight: .medium))
            }

            VStack(alignment: .leading, spacing: 12) {
                keyField(label: "Anthropic API Key", text: $anthropicKey, provider: "anthropic")
                keyField(label: "OpenAI API Key", text: $openaiKey, provider: "openai")
                keyField(label: "Gemini API Key", text: $geminiKey, provider: "gemini")
            }

            if let msg = saveMessage {
                Text(msg)
                    .font(.system(size: 11, weight: .medium))
                    .foregroundColor(Color(red: 0.20, green: 0.78, blue: 0.35))
                    .transition(.opacity)
            }

            Spacer()
        }
        .padding(16)
        .frame(width: 340, height: 260)
        .onAppear {
            anthropicKey = KeychainService.loadKey(provider: "anthropic") ?? ""
            openaiKey = KeychainService.loadKey(provider: "openai") ?? ""
            geminiKey = KeychainService.loadKey(provider: "gemini") ?? ""
        }
    }

    private func keyField(label: String, text: Binding<String>, provider: String) -> some View {
        VStack(alignment: .leading, spacing: 4) {
            Text(label)
                .font(.system(size: 11, weight: .medium))
                .foregroundColor(.secondary)

            HStack {
                SecureField("sk-...", text: text)
                    .textFieldStyle(.roundedBorder)
                    .font(.system(size: 11, design: .monospaced))

                Button("Save") {
                    _ = KeychainService.saveKey(provider: provider, key: text.wrappedValue)
                    withAnimation {
                        saveMessage = "\(provider.capitalized) key securely saved!"
                    }
                    DispatchQueue.main.asyncAfter(deadline: .now() + 2.0) {
                        withAnimation { saveMessage = nil }
                    }
                }
                .buttonStyle(.borderedProminent)
                .controlSize(.small)
            }
        }
    }
}
