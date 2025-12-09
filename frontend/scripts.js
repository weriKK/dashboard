    // Configuration
    const API_BASE = window.location.origin;
    const FEED_COUNT_KEY = 'feedItemCounts';
    const FEED_ORDER_KEY = 'feedOrder';
    const THEME_KEY = 'dashboardTheme';
    let lastDashboardData = null;

    // Theme switching
    function applyTheme(theme) {
      if (theme === 'warm-neutral') {
        document.documentElement.removeAttribute('data-theme');
      } else {
        document.documentElement.setAttribute('data-theme', theme);
      }
      localStorage.setItem(THEME_KEY, theme);
    }

    // Load saved theme on page load
    const savedTheme = localStorage.getItem(THEME_KEY) || 'warm-neutral';
    applyTheme(savedTheme);
    const themeSelect = document.getElementById('themeSelect');
    if (themeSelect) {
      themeSelect.value = savedTheme;
      themeSelect.addEventListener('change', (e) => {
        applyTheme(e.target.value);
      });
    }

    // Helper: feed display preferences
    function loadFeedCountPrefs() {
      try {
        const raw = localStorage.getItem(FEED_COUNT_KEY);
        return raw ? JSON.parse(raw) : {};
      } catch (e) {
        console.warn('Failed to parse feed count prefs', e);
        return {};
      }
    }

    function saveFeedCountPrefs(prefs) {
      try {
        localStorage.setItem(FEED_COUNT_KEY, JSON.stringify(prefs));
      } catch (e) {
        console.warn('Failed to save feed count prefs', e);
      }
    }

    function getFeedCount(feedKey) {
      const prefs = loadFeedCountPrefs();
      const val = prefs[feedKey];
      return Number.isFinite(val) && val > 0 ? val : 10;
    }

    function setFeedCount(feedKey, count) {
      const prefs = loadFeedCountPrefs();
      prefs[feedKey] = count;
      saveFeedCountPrefs(prefs);
    }

    // Feed order and column helpers
    function loadFeedOrder() {
      try {
        const raw = localStorage.getItem(FEED_ORDER_KEY);
        return raw ? JSON.parse(raw) : {};
      } catch (e) {
        console.warn('Failed to parse feed order', e);
        return {};
      }
    }

    function saveFeedOrder(order) {
      try {
        localStorage.setItem(FEED_ORDER_KEY, JSON.stringify(order));
      } catch (e) {
        console.warn('Failed to save feed order', e);
      }
    }

    function normalizeFeedOrder(feedGroups) {
      const existing = loadFeedOrder();
      const keys = feedGroups.map(g => g.source || g.category || 'Feed');
      const result = {};
      for (let i = 0; i < keys.length; i++) {
        const key = keys[i];
        result[key] = existing[key] !== undefined ? existing[key] : [Math.floor(i / Math.ceil(keys.length / 3)), i % Math.ceil(keys.length / 3)];
      }
      if (JSON.stringify(result) !== JSON.stringify(existing)) {
        saveFeedOrder(result);
      }
      return result;
    }

    // Helper: format price change
    function formatChange(change, changePercent) {
      const sign = change >= 0 ? '+' : '';
      const color = change >= 0 ? '#4aa3ff' : '#ff5d5d';
      return `<span style="color: ${color};">(${sign}${change.toFixed(2)} ${sign}${changePercent.toFixed(2)}%)</span>`;
    }

    // Helper: humanize age
    function humanizeAge(isoString) {
      const date = new Date(isoString);
      const now = new Date();
      const diff = now - date;

      const minutes = Math.floor(diff / 60000);
      if (minutes < 60) return `${minutes}m ago`;

      const hours = Math.floor(diff / 3600000);
      if (hours < 24) return `${hours}h ago`;

      const days = Math.floor(diff / 86400000);
      return `${days}d ago`;
    }

    // Helper: get trend bars for stock data
    function getTrendBars(count = 7) {
      // Generate mock trend bars (in real implementation, this would come from API)
      const bars = [];
      for (let i = 0; i < count; i++) {
        const height = Math.floor(Math.random() * 34) + 10;
        const isDown = Math.random() > 0.7;
        bars.push({ height, isDown });
      }
      return bars;
    }

    // Render market card
    function renderMarketCard(stocks, timezones) {
      const tickers = stocks
        .map(stock => {
          const trendBars = getTrendBars(7);
          const barsHtml = trendBars
            .map(
              bar =>
                `<div class="bar${bar.isDown ? ' down' : ''}" style="height: ${bar.height}px"></div>`
            )
            .join('');

          return `
        <div class="ticker">
          <div class="ticker-title">
            <strong>${stock.symbol}</strong>
          </div>
          <div class="ticker-row">
            <div class="trend">
              ${barsHtml}
            </div>
            <div class="value">
              <div class="val">${stock.price > 0 ? stock.price.toFixed(2) : '—'}</div>
              ${stock.price > 0 && stock.changePercent ? formatChange(stock.change, stock.changePercent) : '<div class="chg">(no data)</div>'}
            </div>
          </div>
        </div>
      `;
        })
        .join('');

      const zonesList = (timezones || [])
        .map(tz => {
          if (!tz || !tz.name || !tz.city) {
            return '';
          }
          // Parse offset: could be "-8", "+5:30", etc.
          const offsetStr = String(tz.offset || '0');
          let offsetHours = 0;
          
          if (offsetStr.includes(':')) {
            const parts = offsetStr.split(':');
            const hours = parseInt(parts[0]);
            const mins = parseInt(parts[1]);
            offsetHours = hours + (mins / 60) * (hours >= 0 ? 1 : -1);
          } else {
            offsetHours = parseFloat(offsetStr);
          }
          
          const now = new Date();
          const utcHours = now.getUTCHours();
          const localHours = Math.floor((utcHours + offsetHours + 24) % 24);
          const minutes = String(now.getUTCMinutes()).padStart(2, '0');
          
          return `
        <div class="zone">
          <div class="zone-title">${tz.name}</div>
          <div class="zone-city">${tz.city} (UTC${offsetStr})</div>
          <div class="zone-time">${String(localHours).padStart(2, '0')}:${minutes}</div>
        </div>
      `;
        })
        .filter(x => x)
        .join('');

      return `
    <div class="card market-card">
      <div class="card-header">
        <span>Markets</span>
        <span class="muted">Live data</span>
      </div>
      <div class="market-tickers">
        ${tickers}
      </div>
      <div class="zones">
        ${zonesList}
      </div>
    </div>
  `;
    }

    // Render recommendations card
    function renderRecommendationsCard(recommendations, colorBySource = {}) {
      const items = recommendations
        .map(item => {
          const categoryKey = item.category || item.source || item.feed || '';
          const catColor = item.color || item.categoryColor || colorBySource[categoryKey] || '#7dc3ff';
          const rankGrad = `linear-gradient(to bottom, ${catColor}55, ${catColor})`;
          const sourceLabel = item.sourceName || item.source || item.category || item.feed || '';
          const metaLabel = sourceLabel ? `${item.age} · ${sourceLabel}` : item.age;
          return `
        <div class="item" style="--rankGrad: ${rankGrad}; --itemAccent: ${catColor}; --itemAccentSoft: ${catColor}26; --itemAccentStrong: ${catColor}44" data-link="${item.link}">
          <div class="rank-bar"></div>
          <div>
            <a href="${item.link}" target="_blank" rel="noopener noreferrer">${item.title}</a>
          </div>
          <div class="meta">
            <span>${metaLabel}</span>
            <span class="score">${Math.round(item.score * 100)}%<div class="score-tooltip">${item.reason}</div></span>
          </div>
        </div>
      `;
        })
        .join('');

      return `
    <div class="card recs-card span-2">
      <div class="recs-header">
        <span>Recommended For You</span>
        <div class="recs-actions">
          <span class="muted">ML-ranked</span>
          <span class="settings-icon" id="settingsIcon">⚙️</span>
        </div>
      </div>
      <div class="recs-items">
        ${items}
      </div>
    </div>
  `;
    }

    // Render feed card
    function renderFeedCard(feedKey, categoryName, categoryColor, items, siteUrl, itemCount) {
      const feedItems = items
        .slice(0, itemCount)
        .map(item => {
          const ageMs = new Date() - new Date(item.publishedAt);
          const isOld = ageMs > 24 * 60 * 60 * 1000;
          const rankGrad = `linear-gradient(to bottom, ${categoryColor}55, ${categoryColor})`;
          return `
        <div class="item${isOld ? ' old' : ''}" style="--rankGrad: ${rankGrad}; --itemAccent: ${categoryColor}; --itemAccentSoft: ${categoryColor}26; --itemAccentStrong: ${categoryColor}44" data-link="${item.link}">
          <div class="rank-bar"></div>
          <div class="item-header">
            <div class="item-title">
              <a href="${item.link}" target="_blank" rel="noopener noreferrer">${item.title}</a>
            </div>
            <span class="age">${humanizeAge(item.publishedAt)}</span>
          </div>
        </div>
      `;
        })
        .join('');

      return `
      <div class="card feed-card" style="--accent: ${categoryColor}" data-feed-name="${feedKey}" data-column="0">
      <div class="card-header" draggable="true">
        <div>
          ${siteUrl ? `<a class="card-link" href="${siteUrl}" target="_blank" rel="noopener noreferrer">${categoryName}</a>` : `<span>${categoryName}</span>`}
        </div>
        <div>
          <select class="feed-count-select" data-feed-key="${feedKey}">
            ${[3, 4, 5, 6, 7, 8, 9, 10, 15, 20].map(n => `<option value="${n}" ${n === itemCount ? 'selected' : ''}>${n} items</option>`).join('')}
          </select>
        </div>
      </div>
      <div class="card-content">
        ${feedItems || '<div class="loading">Loading feed...</div>'}
      </div>
    </div>
  `;
    }

    // Track clicks for ML feedback
    function trackClick(event) {
      if (event.target.tagName === 'A' && event.target.href.startsWith('http')) {
        const item = event.target.closest('[data-link]');
        if (item) {
          const link = item.getAttribute('data-link');
          // Send feedback to backend
          (async () => {
            const body = JSON.stringify({
              itemTitle: event.target.textContent,
              itemGUID: link,
              category: 'user_click',
            });
            
            const headers = { 'Content-Type': 'application/json' };
            if (localStorage.getItem('dashboardHmacSecret')) {
              const sig = await signRequest('POST', '/api/feedback', body);
              if (sig) headers['X-HMAC-Signature'] = sig;
            }
            
            fetch(`${API_BASE}/api/feedback`, {
              method: 'POST',
              headers,
              body,
            })
              .then(r => {
                if (!r.ok) console.error('Feedback rejected');
              })
              .catch(err => console.error('Feedback error'));
          })();
        }
      }
    }

    // Handle feed count changes
    function handleFeedCountChange(event) {
      const select = event.target.closest('.feed-count-select');
      if (!select) return;
      const feedKey = select.getAttribute('data-feed-key');
      const count = parseInt(select.value, 10);
      if (!feedKey || !Number.isFinite(count)) return;
      setFeedCount(feedKey, count);
      if (lastDashboardData) {
        renderLayout(lastDashboardData);
      }
    }

    // Drag-and-drop feed reordering across columns
    let dragFeedName = null;
    let dragFromCol = null;

    function showDropZones() {
      document.querySelectorAll('.drop-zone').forEach(z => z.classList.add('active'));
    }

    function hideDropZones() {
      document.querySelectorAll('.drop-zone').forEach(z => {
        z.classList.remove('active', 'drag-over');
      });
    }

    function handleFeedDragStart(event) {
      const header = event.target.closest('.card-header');
      if (!header) return;
      const card = header.closest('.feed-card');
      if (!card) return;
      dragFeedName = card.getAttribute('data-feed-name');
      dragFromCol = parseInt(card.getAttribute('data-column'), 10) || 0;
      event.dataTransfer.effectAllowed = 'move';
      event.dataTransfer.setData('text/plain', dragFeedName);
      card.classList.add('dragging');
      showDropZones();
    }

    function handleFeedDragEnd(event) {
      const card = event.target.closest('.feed-card');
      if (card) card.classList.remove('dragging');
      hideDropZones();
      dragFeedName = null;
      dragFromCol = null;
    }

    function handleFeedDragOver(event) {
      const zone = event.target.closest('.drop-zone');
      const card = event.target.closest('.feed-card');
      if (zone) {
        event.preventDefault();
        event.dataTransfer.dropEffect = 'move';
        zone.classList.add('drag-over');
      } else if (card && card.getAttribute('data-feed-name') !== dragFeedName) {
        event.preventDefault();
        event.dataTransfer.dropEffect = 'move';
        card.classList.add('drag-over');
      }
    }

    function handleFeedDragLeave(event) {
      const zone = event.target.closest('.drop-zone');
      const card = event.target.closest('.feed-card');
      if (zone) zone.classList.remove('drag-over');
      if (card) card.classList.remove('drag-over');
    }

    function handleFeedDrop(event) {
      const zone = event.target.closest('.drop-zone');
      const card = event.target.closest('.feed-card');
      if (!dragFeedName) return;
      event.preventDefault();

      const order = normalizeFeedOrder(lastDashboardData?.feeds || []);
      const feedGroups = lastDashboardData?.feeds || [];

      if (zone) {
        // Drop on drop-zone: move to target column, at the end
        const targetCol = parseInt(zone.getAttribute('data-column'), 10);
        if (Number.isFinite(targetCol)) {
          const colFeeds = feedGroups.filter(f => {
            const name = f.source || f.category || 'Feed';
            const pos = order[name] || [0, 0];
            return pos[0] === targetCol;
          });
          const maxPos = Math.max(...colFeeds.map(f => {
            const name = f.source || f.category || 'Feed';
            return order[name][1];
          }), -1);
          order[dragFeedName] = [targetCol, maxPos + 1];
          console.log('Drop on zone:', dragFeedName, 'moved to col', targetCol, 'pos', maxPos + 1);
          saveFeedOrder(order);
          if (lastDashboardData) renderLayout(lastDashboardData);
        }
      } else if (card) {
        // Drop on card: insert before target (reorder within or across columns)
        const targetName = card.getAttribute('data-feed-name');
        if (targetName && targetName !== dragFeedName) {
          const targetPos = order[targetName] || [0, 0];
          const targetCol = targetPos[0];
          
          // Find all feeds in target column and shift positions
          const colFeeds = feedGroups
            .map(f => f.source || f.category || 'Feed')
            .filter(name => {
              const pos = order[name] || [0, 0];
              return pos[0] === targetCol;
            })
            .sort((a, b) => (order[a][1] || 0) - (order[b][1] || 0));
          
          // Shift all feeds at or after target position
          for (const name of colFeeds) {
            if (name !== dragFeedName) {
              const pos = order[name] || [targetCol, 0];
              if (pos[1] >= targetPos[1]) {
                order[name] = [targetCol, pos[1] + 1];
              }
            }
          }
          
          order[dragFeedName] = [targetCol, targetPos[1]];
          console.log('Drop on card:', dragFeedName, 'moved before', targetName, 'to col', targetCol, 'pos', targetPos[1]);
          saveFeedOrder(order);
          if (lastDashboardData) renderLayout(lastDashboardData);
        }
      }
    }

    // Render layout from data (uses cached data on preference changes)
    function renderLayout(data) {
      if (!data) return;

      const feedGroups = data.feeds || [];
      const colorBySource = {};
      for (const group of feedGroups) {
        const name = group.source || group.category || 'Feed';
        const color = group.color || '#4ba6cd';
        colorBySource[name] = color;
      }

      const feedOrder = normalizeFeedOrder(feedGroups);
      const columns = [[], [], []];
      
      for (const group of feedGroups) {
        const name = group.source || group.category || 'Feed';
        let pos = feedOrder[name];
        if (!Array.isArray(pos) || pos.length < 2) {
          // Initialize if missing or invalid
          feedOrder[name] = [0, 0];
          pos = feedOrder[name];
        }
        const colIndex = Math.max(0, Math.min(2, pos[0]));
        if (!columns[colIndex]) columns[colIndex] = [];
        columns[colIndex].push({ group, position: pos[1] });
      }

      // Sort within each column by position
      for (let i = 0; i < 3; i++) {
        if (columns[i]) {
          columns[i].sort((a, b) => a.position - b.position);
        }
      }

      let html = '';

      // Top grid (market + recommendations)
      html += '<div class="top-grid">';
      html += renderMarketCard(data.stocks || [], data.timezones || []);
      html += renderRecommendationsCard(data.recommendations || [], colorBySource);
      html += '</div>';

      // Feed cards organized by column with drop-zones
      html += '<div class="feeds-masonry">';
      for (let colIndex = 0; colIndex < 3; colIndex++) {
        html += `<div class="feed-column" data-column="${colIndex}">`;
        if (columns[colIndex]) {
          for (const item of columns[colIndex]) {
            const group = item.group;
            const items = group.items || [];
            const name = group.source || group.category || 'Feed';
            const color = group.color || '#4ba6cd';
            const feedKey = group.source || name;
            const itemCount = getFeedCount(feedKey);
            const feedHtml = renderFeedCard(feedKey, name, color, items, group.siteUrl, itemCount);
            const feedWithCol = feedHtml.replace(/data-column="0"/, `data-column="${colIndex}"`);
            html += feedWithCol;
          }
        }
        html += `<div class="drop-zone" data-column="${colIndex}">Drop here</div>`;
        html += '</div>';
      }
      html += '</div>';

      // Handle empty state
      if (!html.includes('card')) {
        html = '<div class="loading">No data available yet. Please check your API key and feed URLs.</div>';
      }

      layout.innerHTML = html;
    }

    // Main render function (fetch + render)
    async function renderDashboard() {
      try {
        const response = await fetch(`${API_BASE}/api/dashboard`);
        if (!response.ok) throw new Error(`HTTP ${response.status}`);

        const data = await response.json();

        lastDashboardData = data;
        renderLayout(data);
        
        // Remove any existing error overlay on success
        const existingError = document.getElementById('errorOverlay');
        if (existingError) existingError.remove();
      } catch (error) {
        console.error('Failed to load dashboard:', error);
        
        // Show error overlay but keep old data visible
        const existingError = document.getElementById('errorOverlay');
        if (existingError) existingError.remove();
        
        const errorOverlay = document.createElement('div');
        errorOverlay.id = 'errorOverlay';
        errorOverlay.className = 'error-overlay';
        errorOverlay.innerHTML = `
          <div class="error-overlay-header">
            <div class="error-overlay-title">⚠️ Failed to refresh dashboard</div>
            <button class="error-overlay-close" onclick="this.parentElement.parentElement.remove()">×</button>
          </div>
          <div class="error-overlay-content">
            <strong>${error.message}</strong><br/>
            ${lastDashboardData ? 'Showing cached data. Will retry in 5 minutes.' : 'Please check:'}
            ${!lastDashboardData ? `
            <ul>
              <li>Backend server is running on port 8080</li>
              <li>config.yaml is properly configured</li>
              <li>Finnhub API key is set (if using stock data)</li>
              <li>RSS feed URLs are accessible</li>
            </ul>
            ` : ''}
          </div>
        `;
        document.body.appendChild(errorOverlay);
        
        // Auto-dismiss after 10 seconds if there's cached data
        if (lastDashboardData) {
          setTimeout(() => {
            if (errorOverlay.parentElement) errorOverlay.remove();
          }, 10000);
        }
      }
    }

    const layout = document.getElementById('layout');

    // Event listeners (set once)
    layout.addEventListener('click', trackClick);
    layout.addEventListener('change', handleFeedCountChange);
    layout.addEventListener('dragstart', handleFeedDragStart);
    layout.addEventListener('dragend', handleFeedDragEnd);
    layout.addEventListener('dragover', handleFeedDragOver);
    layout.addEventListener('dragleave', handleFeedDragLeave);
    layout.addEventListener('drop', handleFeedDrop);

    // Update clocks every minute (without full refresh)
    function updateClocks() {
      const zonesContainer = document.querySelector('.zones');
      if (!zonesContainer) return;
      
      const zoneElements = zonesContainer.querySelectorAll('.zone');
      zoneElements.forEach(zoneEl => {
        const offsetStr = zoneEl.querySelector('.zone-city').textContent.match(/UTC([+\-][\d:]+)/)?.[1] || '0';
        
        // Parse offset
        let offsetHours = 0;
        if (offsetStr.includes(':')) {
          const parts = offsetStr.split(':');
          const hours = parseInt(parts[0]);
          const mins = parseInt(parts[1]);
          offsetHours = hours + (mins / 60) * (hours >= 0 ? 1 : -1);
        } else {
          offsetHours = parseFloat(offsetStr);
        }
        
        const now = new Date();
        const utcHours = now.getUTCHours();
        const localHours = Math.floor((utcHours + offsetHours + 24) % 24);
        const minutes = String(now.getUTCMinutes()).padStart(2, '0');
        
        const timeEl = zoneEl.querySelector('.zone-time');
        if (timeEl) {
          timeEl.textContent = `${String(localHours).padStart(2, '0')}:${minutes}`;
        }
      });
    }

    // HMAC signing utility for protected endpoints
    const hmacSecret = localStorage.getItem('dashboardHmacSecret');
    
    // Settings panel (Ctrl+Shift+S)
    const settingsPanel = document.getElementById('settingsPanel');
    const hmacInput = document.getElementById('hmacInput');
    const settingsStatus = document.getElementById('settingsStatus');
    const saveBtnSettings = document.getElementById('saveBtnSettings');
    const clearBtnSettings = document.getElementById('clearBtnSettings');
    const closeBtnSettings = document.getElementById('closeBtnSettings');

    // Load current HMAC secret into input
    if (hmacSecret) {
      hmacInput.value = hmacSecret;
    }

    // Toggle settings panel by clicking settings icon
    document.addEventListener('click', (e) => {
      if (e.target.classList && e.target.classList.contains('settings-icon')) {
        e.preventDefault();
        settingsPanel.classList.toggle('visible');
        if (settingsPanel.classList.contains('visible')) {
          hmacInput.focus();
          // sync theme select with current theme
          const currentTheme = localStorage.getItem(THEME_KEY) || 'warm-neutral';
          themeSelect.value = currentTheme;
        }
      }
    });

    // Save HMAC secret
    saveBtnSettings.addEventListener('click', () => {
      const secret = hmacInput.value.trim();
      if (!secret) {
        settingsStatus.textContent = '❌ Secret cannot be empty';
        settingsStatus.style.color = '#ff6b6b';
        return;
      }
      localStorage.setItem('dashboardHmacSecret', secret);
      settingsStatus.textContent = '✓ HMAC secret saved';
      settingsStatus.style.color = '#9ee7ff';
      setTimeout(() => {
        settingsStatus.textContent = '';
      }, 2000);
    });

    // Clear HMAC secret
    clearBtnSettings.addEventListener('click', () => {
      localStorage.removeItem('dashboardHmacSecret');
      hmacInput.value = '';
      settingsStatus.textContent = '✓ HMAC secret cleared';
      settingsStatus.style.color = '#9ee7ff';
      setTimeout(() => {
        settingsStatus.textContent = '';
      }, 2000);
    });

    // Close panel
    closeBtnSettings.addEventListener('click', () => {
      settingsPanel.classList.remove('visible');
    });

    // Close panel on Escape
    document.addEventListener('keydown', (e) => {
      if (e.code === 'Escape' && settingsPanel.classList.contains('visible')) {
        settingsPanel.classList.remove('visible');
      }
    });
    
    async function signRequest(method, path, body = '') {
      const secret = localStorage.getItem('dashboardHmacSecret'); // Re-read from localStorage
      if (!secret) return null;
      
      const timestamp = Math.floor(Date.now() / 1000);
      const payload = `${method}|${path}|${timestamp}|${body}`;
      
      const encoder = new TextEncoder();
      const key = await crypto.subtle.importKey('raw', encoder.encode(secret), { name: 'HMAC', hash: 'SHA-256' }, false, ['sign']);
      const signature = await crypto.subtle.sign('HMAC', key, encoder.encode(payload));
      const hexSig = Array.from(new Uint8Array(signature)).map(b => b.toString(16).padStart(2, '0')).join('');
      
      return `${timestamp}:${hexSig}`;
    }

    // Initial load and auto-refresh every 5 minutes
    renderDashboard();
    setInterval(renderDashboard, 5 * 60 * 1000);
    
    // Update clocks every minute
    setInterval(updateClocks, 60 * 1000);