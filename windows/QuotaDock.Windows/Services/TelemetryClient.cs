using System;
using System.Net.Http;
using System.Text.Json;
using System.Threading;
using System.Threading.Tasks;

namespace QuotaDock.Windows.Services
{
    public class TelemetryClient
    {
        private readonly HttpClient _httpClient;
        private CancellationTokenSource? _cts;

        public event Action<string>? TelemetryReceived;

        public TelemetryClient()
        {
            _httpClient = new HttpClient
            {
                BaseAddress = new Uri("http://127.0.0.1:8080/"),
                Timeout = TimeSpan.FromSeconds(5)
            };
        }

        public void Start()
        {
            _cts = new CancellationTokenSource();
            _ = PollTelemetryLoopAsync(_cts.Token);
        }

        public void Stop()
        {
            _cts?.Cancel();
            _cts?.Dispose();
        }

        private async Task PollTelemetryLoopAsync(CancellationToken token)
        {
            while (!token.IsCancellationRequested)
            {
                try
                {
                    var response = await _httpClient.GetStringAsync("quotadock/telemetry", token);
                    TelemetryReceived?.Invoke(response);
                }
                catch
                {
                    // Daemon may be starting or offline
                }

                await Task.Delay(2000, token);
            }
        }
    }
}
