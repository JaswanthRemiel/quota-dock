import SwiftUI

@main
struct QuotaDockApp: App {
    @StateObject private var store = QuotaStore()

    var body: some Scene {
        MenuBarExtra {
            MenuBarView(store: store)
        } label: {
            HStack(spacing: 4) {
                Image(systemName: "gauge.with.needle")
                if store.overallHealthStatus != "healthy" {
                    Circle()
                        .fill(Color(red: 1.0, green: 0.65, blue: 0.0))
                        .frame(width: 5, height: 5)
                }
            }
        }
        .menuBarExtraStyle(.window)
    }
}
