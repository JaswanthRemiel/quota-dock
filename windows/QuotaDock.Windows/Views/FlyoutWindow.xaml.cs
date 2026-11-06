using System;
using System.Diagnostics;
using System.Windows;
using QuotaDock.Windows.Services;

namespace QuotaDock.Windows.Views
{
    public partial class FlyoutWindow : Window
    {
        private readonly TelemetryClient _telemetryClient;

        public FlyoutWindow(TelemetryClient telemetryClient)
        {
            InitializeComponent();
            _telemetryClient = telemetryClient;
        }

        private void CloseButton_Click(object sender, RoutedEventArgs e)
        {
            Hide();
        }

        private void Window_Deactivated(object sender, EventArgs e)
        {
            // Auto-hide when user clicks outside the flyout window (standard Windows 11 tray behavior)
            Hide();
        }

        private void DashboardButton_Click(object sender, RoutedEventArgs e)
        {
            try
            {
                Process.Start(new ProcessStartInfo
                {
                    FileName = "http://127.0.0.1:8080/dashboard",
                    UseShellExecute = true
                });
            }
            catch (Exception ex)
            {
                MessageBox.Show($"Unable to launch dashboard: {ex.Message}", "Error", MessageBoxButton.OK, MessageBoxImage.Error);
            }
        }
    }
}
