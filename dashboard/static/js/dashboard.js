console.log('=== DEBUG: dashboard.js is loading ===');

// Global variables
let metricsHistory = [];
let healthHistory = [];
let autoRefreshInterval = null;

// Load Metrics Data
async function loadMetrics() {
    console.log('=== loadMetrics called ===');
    try {
        const response = await fetch('/metrics');

        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }

        const data = await response.json();
        console.log('DEBUG: Received metrics data:', data);
        console.log('DEBUG: data.system exists?', !!data.system);
        console.log('DEBUG: typeof data:', typeof data);

        // Validate data structure
        if (!data || typeof data !== 'object') {
            throw new Error('Invalid data format received');
        }

        // Validate that system exists
        if (!data.system) {
            console.error('ERROR: data.system is missing!', data);
            console.error('ERROR: Available keys:', Object.keys(data));
            throw new Error('System metrics missing from response');
        }

        // Add to history
        const entry = {
            timestamp: new Date().toISOString(),
            data: data
        };
        metricsHistory.unshift(entry);

        // Keep only last 50 entries
        if (metricsHistory.length > 50) {
            metricsHistory = metricsHistory.slice(0, 50);
        }

        localStorage.setItem('metricsHistory', JSON.stringify(metricsHistory));

        console.log('DEBUG: About to call displayMetrics with:', data);
        displayMetrics(data);
        updateMetricsHistory();

        // Clear any error messages
        const errorDiv = document.querySelector('.error');
        if (errorDiv) {
            errorDiv.remove();
        }

    } catch (error) {
        console.error('ERROR in loadMetrics:', error);
        console.error('ERROR stack:', error.stack);
        const content = document.getElementById('metrics-content');
        if (content) {
            content.innerHTML = '<div class="error">Error loading metrics: ' + error.message + '</div>';
        }
    }
}

// Load Health Data
async function loadHealth() {
    console.log('=== loadHealth called ===');
    try {
        const response = await fetch('/health');
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }

        const data = await response.json();

        // Add to history
        const entry = {
            timestamp: new Date().toISOString(),
            data: data
        };
        healthHistory.unshift(entry);

        // Keep only last 50 entries
        if (healthHistory.length > 50) {
            healthHistory = healthHistory.slice(0, 50);
        }

        localStorage.setItem('healthHistory', JSON.stringify(healthHistory));
        displayHealth(data);
        updateHealthHistory();
    } catch (error) {
        console.error('Error loading health:', error);
        const content = document.getElementById('health-content');
        if (content) {
            content.innerHTML = '<div class="error">Error loading health: ' + error.message + '</div>';
        }
    }
}

// Display Health Data
function displayHealth(data) {
    console.log('=== displayHealth called ===');
    const content = document.getElementById('health-content');
    if (!content) return;

    let html = `
        <div class="current-health">
            <h3>System Health Monitor</h3>
            <div class="timestamp">
                Last updated: ${new Date().toLocaleString()}
            </div>
        </div>

        ${formatUserFriendlyHealth(data)}

        <div class="json-container">
            <h4>Technical Details</h4>
            <div class="json-viewer">
                ${formatJSON(data)}
            </div>
        </div>

        <div class="history-section">
            <h4>Health History</h4>
            <div id="health-history"></div>
        </div>
    `;

    content.innerHTML = html;
}

// Display Metrics Data
function displayMetrics(data) {
    console.log('=== displayMetrics called with:', data);
    const content = document.getElementById('metrics-content');
    if (!content) return;

    // Additional safety check
    if (!data || !data.system) {
        console.error('ERROR: Invalid data passed to displayMetrics:', data);
        content.innerHTML = '<div class="error">Invalid metrics data structure</div>';
        return;
    }

    let html = `
        <div class="current-metrics">
            <h3>System Analytics Dashboard</h3>
            <div class="timestamp">
                Last updated: ${new Date().toLocaleString()}
            </div>
        </div>

        ${formatUserFriendlyMetrics(data)}

        <div class="json-container">
            <h4>Technical Details</h4>
            <div class="json-viewer">
                ${formatJSON(data)}
            </div>
        </div>

        <div class="history-section">
            <h4>Metrics History</h4>
            <div id="metrics-history"></div>
        </div>
    `;

    content.innerHTML = html;
}

// Format Health Data for Display
// Funktion formatUserFriendlyHealth korrigieren:

function formatUserFriendlyHealth(health) {
    let html = '<div class="metrics-grid">';

    // System Status
    const systemStatus = health.healthy ? 'OPERATIONAL' : 'DEGRADED';
    const statusClass = health.healthy ? 'status-good' : 'status-bad';

    html += `
        <div class="metrics-card">
            <h3>Overall System Status</h3>
            <div class="system-status">
                <div class="status-indicator ${statusClass}">
                    <div class="status-dot"></div>
                    <span class="status-text">${systemStatus}</span>
                </div>
                <div class="status-description">Current system state</div>
            </div>
        </div>
    `;

    // System Issues
    const issueCount = health.issues ? health.issues.length : 0;
    html += `
        <div class="user-health-card">
            <div class="health-info">
                <h4>System Issues</h4>
                <div class="health-value ${issueCount > 0 ? 'health-bad' : 'health-good'}">${issueCount}</div>
                <div class="health-description">${issueCount > 0 ? 'Issues detected' : 'No issues detected'}</div>
            </div>
        </div>
    `;

    html += '<div class="performance-section"><h3>Performance Health</h3>';

    // Active Users
    const activeUsers = health.metrics?.system?.active_connections || 0;
    html += `
        <div class="user-health-card">
            <div class="health-info">
                <h4>Active Users</h4>
                <div class="health-value">${activeUsers}</div>
                <div class="health-description">Connected users</div>
            </div>
        </div>
    `;

    // Error Rate - KORRIGIERT
    let errorRateDisplay = '0.0';
    if (health.metrics?.performance) {
        const uploadCount = health.metrics.performance.analysis_count?.upload || 0;
        const uploadErrors = health.metrics.performance.error_count?.upload || 0;
        if (uploadCount > 0) {
            errorRateDisplay = ((uploadErrors / uploadCount) * 100).toFixed(1);
        }
    }

    html += `
        <div class="user-health-card">
            <div class="health-info">
                <h4>Error Rate</h4>
                <div class="health-value">${errorRateDisplay}%</div>
                <div class="health-description">System error percentage</div>
            </div>
        </div>
    `;

    // Average Response Time - KORRIGIERT
    const avgResponseTime = health.average_response_ms || 0;
    html += `
        <div class="user-health-card">
            <div class="health-info">
                <h4>Avg Response Time</h4>
                <div class="health-value">${avgResponseTime}ms</div>
                <div class="health-description">System response time</div>
            </div>
        </div>
    `;

    html += '</div></div>';
    return html;
}

function formatUserFriendlyMetrics(metrics) {
    console.log('=== formatUserFriendlyMetrics called with:', metrics);

    if (!metrics) {
        console.warn('No metrics data provided');
        return '<div class="error">No metrics data available</div>';
    }

    let html = '<div class="metrics-dashboard">';

    // System Overview
    html += '<div class="metrics-section">';
    html += '<h3>System Overview</h3>';
    html += '<div class="metrics-cards">';

    // Add null checks for all nested properties
    const system = metrics.system || {};
    const overall = metrics.overall || {};

    console.log('System data:', system);
    console.log('Overall data:', overall);

    html += `
        <div class="user-metrics-card">
            <div class="metrics-info">
                <h4>Active Users</h4>
                <div class="metrics-value">${system.active_connections || 0}</div>
                <div class="metrics-description">Currently using the system</div>
            </div>
        </div>
        <div class="user-metrics-card">
            <div class="metrics-info">
                <h4>Total Analyses</h4>
                <div class="metrics-value">${overall.total_analyses || 0}</div>
                <div class="metrics-description">Images processed</div>
            </div>
        </div>
        <div class="user-metrics-card">
            <div class="metrics-info">
                <h4>Error Rate</h4>
                <div class="metrics-value">${((overall.overall_error_rate || 0) * 100).toFixed(1)}%</div>
                <div class="metrics-description">System reliability</div>
            </div>
        </div>
    `;

    html += '</div></div>';

    // Cache Performance - add more null checks
    html += '<div class="metrics-section">';
    html += '<h3>Cache Performance</h3>';
    html += '<div class="metrics-cards">';

    const cacheHits = system.cache_hits || 0;
    const cacheMisses = system.cache_misses || 0;
    const totalCache = cacheHits + cacheMisses;
    const hitRate = totalCache > 0 ? (cacheHits / totalCache) : 0;

    html += `
        <div class="user-metrics-card">
            <div class="metrics-info">
                <h4>Cache Hit Rate</h4>
                <div class="metrics-value">${(hitRate * 100).toFixed(1)}%</div>
                <div class="metrics-description">Cache efficiency</div>
            </div>
        </div>
        <div class="user-metrics-card">
            <div class="metrics-info">
                <h4>Cache Hits</h4>
                <div class="metrics-value">${cacheHits}</div>
                <div class="metrics-description">Successful cache retrievals</div>
            </div>
        </div>
        <div class="user-metrics-card">
            <div class="metrics-info">
                <h4>Cache Misses</h4>
                <div class="metrics-value">${cacheMisses}</div>
                <div class="metrics-description">Cache not found</div>
            </div>
        </div>
    `;

    html += '</div></div>';

    // Business Metrics - add null checks
    if (metrics.business) {
        html += '<div class="metrics-section">';
        html += '<h3>AI Detection Analytics</h3>';
        html += '<div class="metrics-cards">';

        const business = metrics.business;

        html += `
            <div class="user-metrics-card">
                <div class="metrics-info">
                    <h4>AI Detection Rate</h4>
                    <div class="metrics-value">${((business.ai_detection_rate || 0) * 100).toFixed(1)}%</div>
                    <div class="metrics-description">Images flagged as AI-generated</div>
                </div>
            </div>
            <div class="user-metrics-card">
                <div class="metrics-info">
                    <h4>Quick Analysis Rate</h4>
                    <div class="metrics-value">${((business.early_exit_rate || 0) * 100).toFixed(1)}%</div>
                    <div class="metrics-description">Images analyzed quickly</div>
                </div>
            </div>
            <div class="user-metrics-card">
                <div class="metrics-info">
                    <h4>AI Detected Count</h4>
                    <div class="metrics-value">${business.ai_detected_count || 0}</div>
                    <div class="metrics-description">Total AI-generated images found</div>
                </div>
            </div>
        `;

        html += '</div></div>';
    }

    // Performance Details - add null checks
    if (metrics.pipeline || metrics.upload) {
        html += '<div class="metrics-section">';
        html += '<h3>Performance Details</h3>';
        html += '<div class="metrics-cards">';

        const pipeline = metrics.pipeline || {};
        const upload = metrics.upload || {};

        html += `
            <div class="user-metrics-card">
                <div class="metrics-info">
                    <h4>Avg Analysis Time</h4>
                    <div class="metrics-value">${pipeline.average_duration || 0}ms</div>
                    <div class="metrics-description">Pipeline processing time</div>
                </div>
            </div>
            <div class="user-metrics-card">
                <div class="metrics-info">
                    <h4>Avg Upload Time</h4>
                    <div class="metrics-value">${upload.average_duration || 0}ms</div>
                    <div class="metrics-description">File upload processing time</div>
                </div>
            </div>
            <div class="user-metrics-card">
                <div class="metrics-info">
                    <h4>Upload Success Rate</h4>
                    <div class="metrics-value">${(((1 - (upload.error_rate || 0)) * 100).toFixed(1))}%</div>
                    <div class="metrics-description">Successful uploads</div>
                </div>
            </div>
        `;

        html += '</div></div>';
    }



    html += '</div></div>';

    // Technical Details
    html += '<div class="metrics-section">';
    html += '<h3>Technical Details</h3>';
    html += '<div class="technical-details">';
    html += '<pre>' + JSON.stringify(metrics, null, 2) + '</pre>';
    html += '</div></div>';

    html += '</div>';

    console.log('=== formatUserFriendlyMetrics completed successfully ===');
    return html;
}

// Update Health History
function updateHealthHistory() {
    const historyDiv = document.getElementById('health-history');
    if (!historyDiv) return;

    let html = '<div class="history-entries">';
    healthHistory.slice(0, 10).forEach(entry => {
        const date = new Date(entry.timestamp);

        const isHealthy = entry.data.healthy !== undefined ? entry.data.healthy : false;
        const hasIssues = entry.data.issues && entry.data.issues.length > 0;

        let status, statusClass;
        if (isHealthy && !hasIssues) {
            status = 'OPERATIONAL';
            statusClass = 'status-success';
        } else if (isHealthy && hasIssues) {
            status = 'DEGRADED';
            statusClass = 'status-warning';
        } else {
            status = 'DOWN';
            statusClass = 'status-error';
        }

        html += `
            <div class="history-entry">
                <span class="history-time">${date.toLocaleTimeString()}</span>
                <span class="history-status ${statusClass}">${status}</span>
                <span class="history-details">${hasIssues ? entry.data.issues.length + ' issues' : 'No issues'}</span>
            </div>
        `;
    });
    html += '</div>';

    historyDiv.innerHTML = html;
}

function updateMetricsHistory() {
    const historyDiv = document.getElementById('metrics-history');
    if (!historyDiv) return;

    let html = '<div class="history-entries">';
    metricsHistory.slice(0, 10).forEach(entry => {
        const date = new Date(entry.timestamp);

        if (!entry.data) {
            console.warn('Invalid entry data:', entry);
            return;
        }

        const errorRate = entry.data.overall?.overall_error_rate || 0;
        const totalAnalyses = entry.data.overall?.total_analyses || 0;

        html += `
            <div class="history-entry">
                <span class="history-time">${date.toLocaleTimeString()}</span>
                <span class="history-details">${totalAnalyses} analyses, ${(errorRate * 100).toFixed(1)}% errors</span>
            </div>
        `;
    });
    html += '</div>';

    historyDiv.innerHTML = html;
}

// Format JSON for display
function formatJSON(data) {
    return JSON.stringify(data, null, 2);
}

function manualRefresh() {
    console.log('=== manualRefresh called ===');
    if (document.getElementById('metrics-content')) {
        loadMetrics();
    }
    if (document.getElementById('health-content')) {
        loadHealth();
    }
}

function toggleAutoRefresh() {
    console.log('=== toggleAutoRefresh called ===');
    const button = document.getElementById('auto-refresh-btn');
    if (autoRefreshInterval) {
        clearInterval(autoRefreshInterval);
        autoRefreshInterval = null;
        if (button) button.textContent = ' Auto-Refresh (10s)';
    } else {
        autoRefreshInterval = setInterval(() => {
            if (document.getElementById('metrics-content')) {
                loadMetrics();
            }
            if (document.getElementById('health-content')) {
                loadHealth();
            }
        }, 10000);
        if (button) button.textContent = '⏸ Stop Auto-Refresh';
    }
}

function exportHistory() {
    console.log('=== exportHistory called ===');
    const data = {
        metrics: metricsHistory,
        health: healthHistory,
        exported: new Date().toISOString()
    };

    const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `system-monitoring-${new Date().toISOString().split('T')[0]}.json`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
}

function clearHistory() {
    console.log('=== clearHistory called ===');
    if (confirm('Are you sure you want to clear all history?')) {
        metricsHistory = [];
        healthHistory = [];
        localStorage.removeItem('metricsHistory');
        localStorage.removeItem('healthHistory');
        updateMetricsHistory();
        updateHealthHistory();
    }
}

// IMPORTANT: Make functions globally available AFTER they are defined
window.manualRefresh = manualRefresh;
window.toggleAutoRefresh = toggleAutoRefresh;
window.exportHistory = exportHistory;
window.clearHistory = clearHistory;
window.loadMetrics = loadMetrics;
window.loadHealth = loadHealth;

// Initialize when page loads
document.addEventListener('DOMContentLoaded', function() {
    console.log('=== Page loaded, initializing ===');

    // Load saved history
    const savedMetricsHistory = localStorage.getItem('metricsHistory');
    if (savedMetricsHistory) {
        try {
            metricsHistory = JSON.parse(savedMetricsHistory);
            console.log('DEBUG: Loaded metrics history entries:', metricsHistory.length);
        } catch (e) {
            console.warn('Could not parse saved metrics history:', e);
            metricsHistory = [];
        }
    }

    const savedHealthHistory = localStorage.getItem('healthHistory');
    if (savedHealthHistory) {
        try {
            healthHistory = JSON.parse(savedHealthHistory);
            console.log('DEBUG: Loaded health history entries:', healthHistory.length);
        } catch (e) {
            console.warn('Could not parse saved health history:', e);
            healthHistory = [];
        }
    }

    // Initial load based on current page
    if (document.getElementById('metrics-content')) {
        console.log('DEBUG: Found metrics-content, loading metrics...');
        loadMetrics();
    }

    if (document.getElementById('health-content')) {
        console.log('DEBUG: Found health-content, loading health...');
        loadHealth();
    }
});

// Test function availability after loading
window.addEventListener('load', function() {
    console.log('=== Page fully loaded, checking functions ===');
    console.log('manualRefresh:', typeof window.manualRefresh);
    console.log('toggleAutoRefresh:', typeof window.toggleAutoRefresh);
    console.log('loadMetrics:', typeof window.loadMetrics);
});

console.log('=== DEBUG: dashboard.js finished loading ===');