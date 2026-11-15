const health = document.querySelector('#health');
const shortenResult = document.querySelector('#shorten-result');
const resolveResult = document.querySelector('#resolve-result');
const chartCanvas = document.querySelector('#metrics-chart');
let chart;

function showResult(target, value, isError = false) {
  target.textContent = typeof value === 'string' ? value : JSON.stringify(value, null, 2);
  target.classList.toggle('error', isError);
}

async function requestJSON(url, options) {
  const response = await fetch(url, options);
  const body = await response.json();
  if (!response.ok) throw new Error(body.error || `Request failed (${response.status})`);
  return body;
}

async function refreshHealth() {
  try {
    const result = await requestJSON('/health');
    health.textContent = `● ${result.status} — PostgreSQL and Redis are reachable`;
    health.className = 'status healthy';
  } catch (error) {
    health.textContent = `● unhealthy — ${error.message}`;
    health.className = 'status unhealthy';
  }
}

async function refreshMetrics() {
  try {
    const data = await requestJSON('/api/v1/metrics');
    const labels = ['shorten', 'resolve', 'redirect', 'cache_hit', 'cache_miss', 'errors'];
    const values = labels.map((label) => data[label]);
    if (!chart) {
      chart = new Chart(chartCanvas, {
        type: 'bar',
        data: { labels, datasets: [{ label: 'Operations', data: values, backgroundColor: '#ff7a59', borderRadius: 8 }] },
        options: { responsive: true, plugins: { legend: { display: false } }, scales: { y: { beginAtZero: true, ticks: { precision: 0 } } } }
      });
    } else {
      chart.data.datasets[0].data = values;
      chart.update();
    }
  } catch (error) {
    console.error(error);
  }
}

document.querySelector('#shorten-form').addEventListener('submit', async (event) => {
  event.preventDefault();
  const form = new FormData(event.currentTarget);
  try {
    const result = await requestJSON('/api/v1/shorten', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ url: form.get('url'), mode: form.get('mode') }) });
    showResult(shortenResult, result);
    document.querySelector('#code').value = result.short_code;
    await refreshMetrics();
  } catch (error) {
    showResult(shortenResult, error.message, true);
  }
});

document.querySelector('#resolve-form').addEventListener('submit', async (event) => {
  event.preventDefault();
  const code = document.querySelector('#code').value.trim();
  try {
    const result = await requestJSON(`/api/v1/resolve/${encodeURIComponent(code)}`);
    showResult(resolveResult, result);
    await refreshMetrics();
  } catch (error) {
    showResult(resolveResult, error.message, true);
  }
});

document.querySelector('#refresh').addEventListener('click', refreshMetrics);
refreshHealth();
refreshMetrics();
setInterval(refreshHealth, 10000);
setInterval(refreshMetrics, 5000);
