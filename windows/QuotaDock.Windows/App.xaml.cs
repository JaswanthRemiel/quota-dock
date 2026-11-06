using System;
using System.Diagnostics;
using System.Drawing;
using System.Threading;
using System.Windows;
using System.Windows.Forms;
using QuotaDock.Windows.Services;
using QuotaDock.Windows.Views;
using Application = System.Windows.Application;

namespace QuotaDock.Windows
{
    public partial class App : Application
    {
        private static Mutex? _mutex;
        private NotifyIcon? _notifyIcon;
        private FlyoutWindow? _flyoutWindow;
        private TelemetryClient? _telemetryClient;

        protected override void OnStartup(StartupEventArgs e)
        {
            const string appName = "QuotaDock_Windows_SingleInstance";
            _mutex = new Mutex(true, appName, out bool createdNew);

            if (!createdNew)
            {
                // App already running
                System.Windows.MessageBox.Show("QuotaDock is already running in your system tray.", "QuotaDock", MessageBoxButton.OK, MessageBoxImage.Information);
                Current.Shutdown();
                return;
            }

            base.OnStartup(e);

            _telemetryClient = new TelemetryClient();
            _flyoutWindow = new FlyoutWindow(_telemetryClient);

            InitializeTrayIcon();
            _telemetryClient.Start();
        }

        private void InitializeTrayIcon()
        {
            _notifyIcon = new NotifyIcon
            {
                Icon = SystemIcons.Shield, // Default icon; can be customized
                Text = "QuotaDock: AI Gateway Active (127.0.0.1:8080)",
                Visible = true
            };

            _notifyIcon.Click += (sender, args) =>
            {
                if (args is MouseEventArgs mouseArgs && mouseArgs.Button == MouseButtons.Left)
                {
                    ToggleFlyout();
                }
            };

            var contextMenu = new ContextMenuStrip();
            contextMenu.Items.Add("Open Dashboard", null, (s, a) => OpenWebDashboard());
            contextMenu.Items.Add("Show Telemetry Flyout", null, (s, a) => ShowFlyout());
            contextMenu.Items.Add(new ToolStripSeparator());
            contextMenu.Items.Add("Exit QuotaDock", null, (s, a) => ExitApplication());

            _notifyIcon.ContextMenuStrip = contextMenu;
        }

        private void ToggleFlyout()
        {
            if (_flyoutWindow == null) return;

            if (_flyoutWindow.IsVisible)
            {
                _flyoutWindow.Hide();
            }
            else
            {
                PositionAndShowFlyout();
            }
        }

        private void ShowFlyout()
        {
            if (_flyoutWindow == null) return;
            PositionAndShowFlyout();
        }

        private void PositionAndShowFlyout()
        {
            if (_flyoutWindow == null) return;

            // Position above system tray in bottom right corner
            var workingArea = Screen.PrimaryScreen?.WorkingArea ?? new Rectangle(0, 0, 1920, 1080);
            _flyoutWindow.Left = workingArea.Right - _flyoutWindow.Width - 16;
            _flyoutWindow.Top = workingArea.Bottom - _flyoutWindow.Height - 16;

            _flyoutWindow.Show();
            _flyoutWindow.Activate();
        }

        private void OpenWebDashboard()
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
                System.Windows.MessageBox.Show($"Unable to open dashboard: {ex.Message}", "Error", MessageBoxButton.OK, MessageBoxImage.Error);
            }
        }

        private void ExitApplication()
        {
            _notifyIcon?.Dispose();
            _telemetryClient?.Stop();
            Current.Shutdown();
        }

        protected override void OnExit(ExitEventArgs e)
        {
            _notifyIcon?.Dispose();
            _telemetryClient?.Stop();
            _mutex?.ReleaseMutex();
            base.OnExit(e);
        }
    }
}
