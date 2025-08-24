class LoadTestDashboard {
    constructor() {
        this.ws = null;
        this.charts = {};
        this.isRunning = false;
        this.statsHistory = [];
        this.maxHistoryPoints = 60;
        
        this.initializeElements();
        this.initializeCharts();
        this.bindEvents();
        // Disable WebSocket for now, use polling only
        this.updateConnectionStatus(false);
        this.loadTestHistory();
    }

    initializeElements() {
        this.elements = {
            startBtn: document.getElementById('startBtn'),
            stopBtn: document.getElementById('stopBtn'),
            connectionStatus: document.getElementById('connectionStatus'),
            testStatus: document.getElementById('testStatus'),
            
            // Metrics
            totalRequests: document.getElementById('totalRequests'),
            successRate: document.getElementById('successRate'),
            rps: document.getElementById('rps'),
            p99Latency: document.getElementById('p99Latency'),
            
            // Config inputs
            testName: document.getElementById('testName'),
            totalOperations: document.getElementById('totalOperations'),
            accountCount: document.getElementById('accountCount'),
            initialBalance: document.getElementById('initialBalance'),
            balanceDisplay: document.getElementById('balanceDisplay'),
            workers: document.getElementById('workers'),
            workersValue: document.getElementById('workersValue'),
            duration: document.getElementById('duration'),
            rampUp: document.getElementById('rampUp'),
            thinkTime: document.getElementById('thinkTime'),
            minAmount: document.getElementById('minAmount'),
            maxAmount: document.getElementById('maxAmount'),
            minAmountDisplay: document.getElementById('minAmountDisplay'),
            maxAmountDisplay: document.getElementById('maxAmountDisplay'),
            scenarioPreset: document.getElementById('scenarioPreset'),
            
            // Operation mix
            depositMix: document.getElementById('depositMix'),
            withdrawMix: document.getElementById('withdrawMix'),
            transferMix: document.getElementById('transferMix'),
            balanceMix: document.getElementById('balanceMix'),
            depositPercent: document.getElementById('depositPercent'),
            withdrawPercent: document.getElementById('withdrawPercent'),
            transferPercent: document.getElementById('transferPercent'),
            balancePercent: document.getElementById('balancePercent'),
            
            // Operation stats
            depositStats: document.getElementById('depositStats'),
            withdrawStats: document.getElementById('withdrawStats'),
            transferStats: document.getElementById('transferStats'),
            balanceStats: document.getElementById('balanceStats'),
            
            // History
            historyTable: document.getElementById('historyTable')
        };
    }

    initializeCharts() {
        // Throughput Chart
        const throughputCtx = document.getElementById('throughputChart').getContext('2d');
        this.charts.throughput = new Chart(throughputCtx, {
            type: 'line',
            data: {
                labels: [],
                datasets: [{
                    label: 'Requests/sec',
                    data: [],
                    borderColor: 'rgb(75, 192, 192)',
                    backgroundColor: 'rgba(75, 192, 192, 0.2)',
                    tension: 0.4
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                scales: {
                    y: {
                        beginAtZero: true
                    }
                },
                plugins: {
                    legend: {
                        display: false
                    }
                }
            }
        });

        // Latency Chart
        const latencyCtx = document.getElementById('latencyChart').getContext('2d');
        this.charts.latency = new Chart(latencyCtx, {
            type: 'line',
            data: {
                labels: [],
                datasets: [{
                    label: 'P99 Latency (ms)',
                    data: [],
                    borderColor: 'rgb(255, 99, 132)',
                    backgroundColor: 'rgba(255, 99, 132, 0.2)',
                    tension: 0.4
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                scales: {
                    y: {
                        beginAtZero: true
                    }
                },
                plugins: {
                    legend: {
                        display: false
                    }
                }
            }
        });
    }

    // Helper functions for formatting
    formatBRL(cents) {
        const reais = cents / 100;
        return new Intl.NumberFormat('pt-BR', {
            style: 'currency',
            currency: 'BRL'
        }).format(reais);
    }
    
    formatNumberPtBR(num) {
        return new Intl.NumberFormat('pt-BR').format(num);
    }
    
    parseBRL(str) {
        // Remove R$, spaces, and convert comma to dot for decimal
        const cleaned = str.replace(/[R$\s.]/g, '').replace(',', '.');
        return Math.round(parseFloat(cleaned) * 100); // Convert to cents
    }
    
    parseFormattedNumber(str) {
        // Remove dots used as thousand separators
        return parseInt(str.replace(/\./g, ''), 10);
    }

    bindEvents() {
        // Initialize Bootstrap tooltips
        const tooltipTriggerList = document.querySelectorAll('[data-bs-toggle="tooltip"]');
        const tooltipList = [...tooltipTriggerList].map(tooltipTriggerEl => new bootstrap.Tooltip(tooltipTriggerEl));
        
        this.elements.startBtn.addEventListener('click', () => this.startTest());
        this.elements.stopBtn.addEventListener('click', () => this.stopTest());
        
        this.elements.workers.addEventListener('input', (e) => {
            this.elements.workersValue.textContent = e.target.value;
        });

        // Handle total operations formatting
        this.elements.totalOperations.addEventListener('input', (e) => {
            // Remove all non-digit characters while typing
            const value = e.target.value.replace(/\D/g, '');
            if (value) {
                e.target.setAttribute('data-value', value);
            }
        });
        
        this.elements.totalOperations.addEventListener('blur', (e) => {
            const value = e.target.getAttribute('data-value') || e.target.value.replace(/\D/g, '');
            if (value) {
                const num = parseInt(value, 10);
                e.target.value = this.formatNumberPtBR(num);
                e.target.setAttribute('data-value', num);
            }
        });
        
        // Handle initial balance formatting
        this.elements.initialBalance.addEventListener('focus', (e) => {
            // Show raw value when focused for easier editing
            const value = e.target.getAttribute('data-value');
            if (value) {
                e.target.value = value;
            }
        });
        
        this.elements.initialBalance.addEventListener('blur', (e) => {
            const value = e.target.value.replace(/\D/g, '');
            if (value) {
                const cents = parseInt(value, 10);
                e.target.value = this.formatBRL(cents);
                e.target.setAttribute('data-value', cents);
            }
        });
        
        // Handle min amount formatting
        this.elements.minAmount.addEventListener('focus', (e) => {
            const value = e.target.getAttribute('data-value');
            if (value) {
                e.target.value = value;
            }
        });
        
        this.elements.minAmount.addEventListener('blur', (e) => {
            const value = e.target.value.replace(/\D/g, '');
            if (value) {
                const cents = parseInt(value, 10);
                e.target.value = this.formatBRL(cents);
                e.target.setAttribute('data-value', cents);
            }
        });
        
        // Handle max amount formatting
        this.elements.maxAmount.addEventListener('focus', (e) => {
            const value = e.target.getAttribute('data-value');
            if (value) {
                e.target.value = value;
            }
        });
        
        this.elements.maxAmount.addEventListener('blur', (e) => {
            const value = e.target.value.replace(/\D/g, '');
            if (value) {
                const cents = parseInt(value, 10);
                e.target.value = this.formatBRL(cents);
                e.target.setAttribute('data-value', cents);
            }
        });

        // Operation mix sliders
        const mixSliders = ['depositMix', 'withdrawMix', 'transferMix', 'balanceMix'];
        const mixLabels = ['depositPercent', 'withdrawPercent', 'transferPercent', 'balancePercent'];
        
        mixSliders.forEach((slider, index) => {
            this.elements[slider].addEventListener('input', () => {
                this.updateOperationMix();
            });
        });

        this.elements.scenarioPreset.addEventListener('change', (e) => {
            this.loadPreset(e.target.value);
        });
    }

    updateOperationMix() {
        const deposit = parseInt(this.elements.depositMix.value);
        const withdraw = parseInt(this.elements.withdrawMix.value);
        const transfer = parseInt(this.elements.transferMix.value);
        const balance = parseInt(this.elements.balanceMix.value);
        
        const total = deposit + withdraw + transfer + balance;
        
        if (total > 0) {
            const depositPct = Math.round((deposit / total) * 100);
            const withdrawPct = Math.round((withdraw / total) * 100);
            const transferPct = Math.round((transfer / total) * 100);
            const balancePct = Math.round((balance / total) * 100);
            
            this.elements.depositPercent.textContent = `${depositPct}%`;
            this.elements.withdrawPercent.textContent = `${withdrawPct}%`;
            this.elements.transferPercent.textContent = `${transferPct}%`;
            this.elements.balancePercent.textContent = `${balancePct}%`;
        }
    }

    loadPreset(preset) {
        const presets = {
            'default': {
                deposit: 25,
                withdraw: 25,
                transfer: 35,
                balance: 15,
                workers: 100,
                accounts: 1000
            },
            'high-concurrency': {
                deposit: 10,
                withdraw: 10,
                transfer: 70,
                balance: 10,
                workers: 200,
                accounts: 100
            },
            'read-heavy': {
                deposit: 5,
                withdraw: 5,
                transfer: 10,
                balance: 80,
                workers: 50,
                accounts: 5000
            }
        };

        if (presets[preset]) {
            const p = presets[preset];
            this.elements.depositMix.value = p.deposit;
            this.elements.withdrawMix.value = p.withdraw;
            this.elements.transferMix.value = p.transfer;
            this.elements.balanceMix.value = p.balance;
            this.elements.workers.value = p.workers;
            this.elements.workersValue.textContent = p.workers;
            this.elements.accountCount.value = p.accounts;
            this.updateOperationMix();
        }
    }

    connectWebSocket() {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/ws/stats`;
        
        try {
            this.ws = new WebSocket(wsUrl);
            
            this.ws.onopen = () => {
                this.updateConnectionStatus(true);
                console.log('WebSocket connected successfully');
            };
            
            this.ws.onclose = () => {
                this.updateConnectionStatus(false);
                console.log('WebSocket disconnected, using polling fallback');
                this.startPolling();
                setTimeout(() => this.connectWebSocket(), 10000);
            };
            
            this.ws.onmessage = (event) => {
                const stats = JSON.parse(event.data);
                this.updateStats(stats);
            };
            
            this.ws.onerror = (error) => {
                console.error('WebSocket error:', error);
                this.updateConnectionStatus(false);
                this.startPolling();
            };
        } catch (error) {
            console.error('WebSocket initialization failed:', error);
            this.updateConnectionStatus(false);
            this.startPolling();
        }
    }
    
    startPolling() {
        if (this.pollingInterval) {
            clearInterval(this.pollingInterval);
        }
        
        // Start polling immediately
        console.log('Starting polling for test status');
        this.pollingInterval = setInterval(() => {
            this.pollTestStatus();
        }, 2000);
    }
    
    async pollTestStatus() {
        try {
            const response = await fetch('/api/test/status');
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            const data = await response.json();
            
            if (data.running && this.isRunning) {
                // Only use real backend data
                if (data.live_stats) {
                    console.log('Using real live_stats:', data.live_stats);
                    this.updateStats(data.live_stats);
                    this.updateConnectionStatus(true);
                } else {
                    console.log('No live_stats available from backend');
                    this.updateConnectionStatus(false);
                }
            } else if (!data.running && this.isRunning) {
                // Test completed, update UI state without recursive call
                console.log('Test completed, updating UI state');
                this.isRunning = false;
                this.updateTestStatus(false);
                
                if (this.pollingInterval) {
                    clearInterval(this.pollingInterval);
                    this.pollingInterval = null;
                    console.log('Stopped polling - test completed');
                }
                
                this.loadTestHistory();
            } else if (!this.isRunning) {
                // Stop polling if UI says no test is running
                if (this.pollingInterval) {
                    clearInterval(this.pollingInterval);
                    this.pollingInterval = null;
                }
            }
        } catch (error) {
            console.error('Error polling test status:', error);
            this.updateConnectionStatus(false);
            // Don't stop polling immediately, maybe it's a temporary network issue
        }
    }

    updateConnectionStatus(connected) {
        const icon = this.elements.connectionStatus.querySelector('i');
        if (connected) {
            icon.className = 'fas fa-circle status-running';
            this.elements.connectionStatus.innerHTML = '<i class="fas fa-circle status-running"></i> Connected';
        } else {
            icon.className = 'fas fa-circle status-idle';
            this.elements.connectionStatus.innerHTML = '<i class="fas fa-circle status-idle"></i> Disconnected';
        }
    }

    async startTest() {
        const deposit = parseInt(this.elements.depositMix.value);
        const withdraw = parseInt(this.elements.withdrawMix.value);
        const transfer = parseInt(this.elements.transferMix.value);
        const balance = parseInt(this.elements.balanceMix.value);
        
        const total = deposit + withdraw + transfer + balance;
        
        const config = {
            name: this.elements.testName.value,
            total_operations: parseInt(this.elements.totalOperations.getAttribute('data-value') || this.parseFormattedNumber(this.elements.totalOperations.value)),
            account_count: parseInt(this.elements.accountCount.value),
            initial_balance: parseInt(this.elements.initialBalance.getAttribute('data-value') || this.parseBRL(this.elements.initialBalance.value)),
            workers: parseInt(this.elements.workers.value),
            duration_seconds: parseInt(this.elements.duration.value),
            ramp_up_seconds: parseInt(this.elements.rampUp.value),
            think_time_ms: parseInt(this.elements.thinkTime.value),
            operation_mix: {
                deposit: deposit / total,
                withdraw: withdraw / total,
                transfer: transfer / total,
                balance: balance / total
            },
            amount_range: {
                min: parseFloat(this.elements.minAmount.getAttribute('data-value') || this.parseBRL(this.elements.minAmount.value)),
                max: parseFloat(this.elements.maxAmount.getAttribute('data-value') || this.parseBRL(this.elements.maxAmount.value))
            }
        };

        try {
            const response = await fetch('/api/test/start', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(config)
            });

            if (response.ok) {
                this.isRunning = true;
                this.updateTestStatus(true);
                this.clearCharts();
                this.statsHistory = [];
                this.currentRequests = 0;
                // Always use polling for now
                this.startPolling();
            } else {
                alert('Failed to start test: ' + await response.text());
            }
        } catch (error) {
            alert('Error starting test: ' + error.message);
        }
    }

    async stopTest() {
        // Prevent multiple stop calls
        if (!this.isRunning) {
            console.log('Test already stopped');
            return;
        }
        
        console.log('Stopping test...');
        
        try {
            const response = await fetch('/api/test/stop', {
                method: 'POST'
            });

            if (response.ok) {
                const data = await response.json();
                console.log('Stop response:', data.status);
            } else {
                console.error('Failed to stop test:', response.status);
            }
        } catch (error) {
            console.error('Error stopping test:', error);
        }
        
        // Always update UI state regardless of API response
        this.isRunning = false;
        this.updateTestStatus(false);
        
        if (this.pollingInterval) {
            clearInterval(this.pollingInterval);
            this.pollingInterval = null;
            console.log('Stopped polling');
        }
        
        this.loadTestHistory();
    }

    updateTestStatus(running) {
        if (running) {
            this.elements.testStatus.className = 'badge bg-success';
            this.elements.testStatus.innerHTML = '<i class="fas fa-circle status-running"></i> Running';
            this.elements.startBtn.disabled = true;
            this.elements.stopBtn.disabled = false;
        } else {
            this.elements.testStatus.className = 'badge bg-secondary';
            this.elements.testStatus.innerHTML = '<i class="fas fa-circle"></i> Idle';
            this.elements.startBtn.disabled = false;
            this.elements.stopBtn.disabled = true;
        }
    }

    updateStats(stats) {
        // Update metric cards with safety checks
        this.elements.totalRequests.textContent = this.formatNumber(stats.total_requests || 0);
        this.elements.successRate.textContent = `${((stats.success_rate || 0) * 100).toFixed(1)}%`;
        this.elements.rps.textContent = (stats.requests_per_second || 0).toFixed(0);
        this.elements.p99Latency.textContent = `${(stats.p99_latency_ms || 0).toFixed(0)}ms`;
        
        // Update operation stats
        if (stats.operations) {
            this.elements.depositStats.textContent = this.formatNumber(stats.operations.deposit?.count || 0);
            this.elements.withdrawStats.textContent = this.formatNumber(stats.operations.withdraw?.count || 0);
            this.elements.transferStats.textContent = this.formatNumber(stats.operations.transfer?.count || 0);
            this.elements.balanceStats.textContent = this.formatNumber(stats.operations.balance?.count || 0);
        }
        
        // Update charts
        this.updateCharts(stats);
    }

    updateCharts(stats) {
        const timestamp = new Date().toLocaleTimeString();
        
        // Add to history
        this.statsHistory.push({
            timestamp: timestamp,
            rps: stats.requests_per_second,
            p99: stats.p99_latency_ms
        });
        
        // Keep only last N points
        if (this.statsHistory.length > this.maxHistoryPoints) {
            this.statsHistory.shift();
        }
        
        // Update throughput chart
        this.charts.throughput.data.labels = this.statsHistory.map(s => s.timestamp);
        this.charts.throughput.data.datasets[0].data = this.statsHistory.map(s => s.rps);
        this.charts.throughput.update('none');
        
        // Update latency chart
        this.charts.latency.data.labels = this.statsHistory.map(s => s.timestamp);
        this.charts.latency.data.datasets[0].data = this.statsHistory.map(s => s.p99);
        this.charts.latency.update('none');
    }

    clearCharts() {
        this.charts.throughput.data.labels = [];
        this.charts.throughput.data.datasets[0].data = [];
        this.charts.throughput.update();
        
        this.charts.latency.data.labels = [];
        this.charts.latency.data.datasets[0].data = [];
        this.charts.latency.update();
    }

    async loadTestHistory() {
        try {
            const response = await fetch('/api/test/history');
            const history = await response.json();
            
            if (history && history.length > 0) {
                const rows = history.map(test => `
                    <tr>
                        <td>${test.id}</td>
                        <td>${new Date(test.start_time).toLocaleString()}</td>
                        <td>${test.duration.toFixed(0)}s</td>
                        <td>
                            <span class="badge ${this.getStatusBadgeClass(test.status)}">
                                ${test.status}
                            </span>
                        </td>
                        <td>${test.throughput.toFixed(0)} ops/s</td>
                        <td>${test.p99_latency.toFixed(0)}ms</td>
                        <td>${(test.success_rate * 100).toFixed(1)}%</td>
                        <td>
                            <button class="btn btn-sm btn-outline-primary" onclick="viewReport('${test.id}')">
                                <i class="fas fa-chart-bar"></i>
                            </button>
                        </td>
                    </tr>
                `).join('');
                
                this.elements.historyTable.innerHTML = rows;
            }
        } catch (error) {
            console.error('Failed to load test history:', error);
        }
    }

    getStatusBadgeClass(status) {
        const classes = {
            'EXCELLENT': 'bg-success',
            'GOOD': 'bg-info',
            'ACCEPTABLE': 'bg-warning',
            'NEEDS_IMPROVEMENT': 'bg-danger'
        };
        return classes[status] || 'bg-secondary';
    }

    formatNumber(num) {
        if (num >= 1000000) {
            return `${(num / 1000000).toFixed(1)}M`;
        } else if (num >= 1000) {
            return `${(num / 1000).toFixed(1)}K`;
        }
        return num.toString();
    }
}

// Initialize dashboard when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    window.dashboard = new LoadTestDashboard();
});

// Helper function to view report
function viewReport(testId) {
    window.open(`/api/test/report/${testId}`, '_blank');
}