    // Configuration
    const API_BASE = window.location.origin;
    const FEED_COUNT_KEY = 'feedItemCounts';
    const FEED_ORDER_KEY = 'feedOrder';
    let lastDashboardData = null;

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

    // Helper: humanize age
    function humanizeAge(isoString) {
      const date = new Date(isoString);
      const now = new Date();
      const diff = now - date;

      const minutes = Math.floor(diff / 60000);
      if (minutes < 60) return `${minutes}m`;

      const hours = Math.floor(diff / 3600000);
      if (hours < 24) return `${hours}h`;

      const days = Math.floor(diff / 86400000);
      return `${days}d`;
    }

    function buildTopRatedMap(topRatedItems, allVisibleLinks) {
      const topRatedMap = {};
      if (!Array.isArray(topRatedItems)) return topRatedMap;

      // Scan ranked items until we find 5 that match visible feed items
      let rank = 0;
      for (let i = 0; i < topRatedItems.length && rank < 5; i++) {
        const item = topRatedItems[i];
        if (!item || !item.link) continue;
        if (topRatedMap[item.link]) continue;
        if (!allVisibleLinks.has(item.link)) continue;
        rank++;
        topRatedMap[item.link] = {
          rank: rank,
          score: Math.round((item.score || 0) * 100),
        };
      }

      return topRatedMap;
    }

    // Render feed card
    function renderFeedCard(feedKey, categoryName, categoryColor, items, siteUrl, itemCount, isMobile = false, topRatedMap = {}) {
      const feedItems = items
        .slice(0, itemCount)
        .map(item => {
          const ageMs = new Date() - new Date(item.publishedAt);
          let ageClass = '';
          if (ageMs > 24 * 60 * 60 * 1000) {
            ageClass = ' day_old';
          } else if (ageMs > 12 * 60 * 60 * 1000) {
            ageClass = ' half_day_old';
          }
          const age = humanizeAge(item.publishedAt);
          const topRated = topRatedMap[item.link];
          const topRatedClass = topRated ? ' top-rated-item' : '';
          const topRatedBadge = topRated
            ? `<span class="top-rated-badge" title="Top ${topRated.rank} rated · ${topRated.score}%">Top ${topRated.rank}</span>`
            : '';
          return `<li class="${topRatedClass.trim()}"><a href="${item.link}" target="_blank" rel="noopener noreferrer" class="${ageClass}">${item.title}</a><div class="item-meta-right">${topRatedBadge}<span class="age">${age}</span></div></li>`;
        })
        .join('');

      const draggableAttr = isMobile ? '' : ' draggable="true"';

      return `
          <div class="card feed-card" style="--accent: ${categoryColor}" data-feed-name="${feedKey}" data-column="0">
          <div class="feed-topline"></div>
          <div class="feed-titlebar"${draggableAttr}>
        <div>
          ${siteUrl ? `<a class="card-link feed-title-label" href="${siteUrl}" target="_blank" rel="noopener noreferrer">${categoryName}</a>` : `<span class="feed-title-label">${categoryName}</span>`}
        </div>
        <div>
          <select class="feed-count-select" data-feed-key="${feedKey}">
            ${[3, 4, 5, 6, 7, 8, 9, 10, 15, 20].map(n => `<option value="${n}" ${n === itemCount ? 'selected' : ''}>${n} items</option>`).join('')}
          </select>
        </div>
      </div>
      <ul class="card-content">
        ${feedItems || '<li class="loading">Loading feed...</li>'}
      </ul>
    </div>
  `;
    }

    function sendFeedback(title, link) {
      const secret = localStorage.getItem('dashboardHmacSecret');
      if (!secret) return;

      const body = JSON.stringify({
        itemTitle: title,
        itemLink: link,
        category: 'user_click',
      });

      (async () => {
        const headers = { 'Content-Type': 'application/json' };
        const sig = await signRequest('POST', '/api/feedback', body);
        if (sig) headers['X-HMAC-Signature'] = sig;

        fetch(`${API_BASE}/api/feedback`, {
          method: 'POST',
          headers,
          body,
          keepalive: true,
        }).catch(() => {});
      })();
    }

    // Track clicks for ML feedback (left-click)
    function trackClick(event) {
      const anchor = event.target.closest('a[href^="http"]');
      if (anchor) {
        sendFeedback(anchor.textContent, anchor.href);
      }
    }

    // Track middle-click for ML feedback
    function trackAuxClick(event) {
      if (event.button === 1) {
        const anchor = event.target.closest('a[href^="http"]');
        if (anchor) {
          sendFeedback(anchor.textContent, anchor.href);
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
      const header = event.target.closest('.feed-titlebar');
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

      const isMobile = window.matchMedia('(max-width: 768px)').matches;

      const feedGroups = data.feeds || [];
      const topRatedItems = data.topRated || [];

      // Collect all visible item links across all feeds so badge mapping
      // can skip ranked items that aren't currently shown on screen
      const allVisibleLinks = new Set();
      for (const group of feedGroups) {
        const feedKey = group.source || group.category || 'Feed';
        const itemCount = getFeedCount(feedKey);
        const items = group.items || [];
        for (let i = 0; i < itemCount && i < items.length; i++) {
          if (items[i].link) allVisibleLinks.add(items[i].link);
        }
      }

      const topRatedMap = buildTopRatedMap(topRatedItems, allVisibleLinks);

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
            const feedHtml = renderFeedCard(feedKey, name, color, items, group.siteUrl, itemCount, isMobile, topRatedMap);
            const feedWithCol = feedHtml.replace(/data-column="0"/, `data-column="${colIndex}"`);
            html += feedWithCol;
          }
        }
        if (!isMobile) {
          html += `<div class="drop-zone" data-column="${colIndex}">Drop here</div>`;
        }
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
    layout.addEventListener('auxclick', trackAuxClick);
    layout.addEventListener('change', handleFeedCountChange);
    layout.addEventListener('dragstart', handleFeedDragStart);
    layout.addEventListener('dragend', handleFeedDragEnd);
    layout.addEventListener('dragover', handleFeedDragOver);
    layout.addEventListener('dragleave', handleFeedDragLeave);
    layout.addEventListener('drop', handleFeedDrop);

    // Re-render on viewport change so drag/drop UI toggles with breakpoints
    const mobileQuery = window.matchMedia('(max-width: 768px)');
    mobileQuery.addEventListener('change', () => {
      if (lastDashboardData) {
        renderLayout(lastDashboardData);
      }
    });



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

    document.getElementById('settingsGear').addEventListener('click', () => {
      settingsPanel.classList.toggle('visible');
      if (settingsPanel.classList.contains('visible')) hmacInput.focus();
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