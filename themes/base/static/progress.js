(function () {
    'use strict';

    var STORAGE_KEY = 'tb-progress';
    var SCHEMA_VERSION = 1;
    var TB_K = '';

    function _zb(s) {
        if (!TB_K) return s;
        var raw = atob(s);
        var keyBytes = new TextEncoder().encode(TB_K);
        var bytes = new Uint8Array(raw.length);
        for (var i = 0; i < raw.length; i++) {
            bytes[i] = raw.charCodeAt(i) ^ keyBytes[i % keyBytes.length];
        }
        return new TextDecoder('utf-8').decode(bytes);
    }

    function _rn(node) {
        if (!node) return;
        var payload = node.getAttribute('data-q');
        if (payload === null) return;
        node.innerHTML = _zb(payload);
        node.removeAttribute('data-q');
        node.removeAttribute('data-z');
    }

    function _rw(root) {
        if (!root) return;
        var nodes = root.querySelectorAll('[data-z]');
        for (var i = 0; i < nodes.length; i++) _rn(nodes[i]);
    }

    function defaultBlob() {
        return { version: SCHEMA_VERSION, completedCampaigns: [], stageProgress: {} };
    }

    function readBlob() {
        try {
            var raw = window.localStorage.getItem(STORAGE_KEY);
            if (!raw) return defaultBlob();
            var parsed = JSON.parse(raw);
            if (!parsed || typeof parsed !== 'object') return defaultBlob();
            if (parsed.version !== SCHEMA_VERSION) return defaultBlob();
            if (!Array.isArray(parsed.completedCampaigns)) parsed.completedCampaigns = [];
            if (!parsed.stageProgress || typeof parsed.stageProgress !== 'object') parsed.stageProgress = {};
            return parsed;
        } catch (e) {
            return defaultBlob();
        }
    }

    function writeBlob(blob) {
        try {
            window.localStorage.setItem(STORAGE_KEY, JSON.stringify(blob));
        } catch (e) {
            /* quota, disabled storage, etc. — silently ignored */
        }
    }

    function stageKey(campaign, stage) {
        return campaign + '/' + stage;
    }

    function defaultEntry() {
        return { solved: false, visited: false, answer: '', hintsRevealed: {} };
    }

    function getStageEntry(campaign, stage) {
        var blob = readBlob();
        var key = stageKey(campaign, stage);
        var entry = blob.stageProgress[key];
        if (!entry || typeof entry !== 'object') return defaultEntry();
        return {
            solved: entry.solved === true,
            visited: entry.visited === true,
            answer: typeof entry.answer === 'string' ? entry.answer : '',
            hintsRevealed: (entry.hintsRevealed && typeof entry.hintsRevealed === 'object') ? entry.hintsRevealed : {}
        };
    }

    function mutateEntry(campaign, stage, fn) {
        var blob = readBlob();
        var key = stageKey(campaign, stage);
        var entry = blob.stageProgress[key];
        if (!entry || typeof entry !== 'object') entry = defaultEntry();
        fn(entry);
        blob.stageProgress[key] = entry;
        writeBlob(blob);
    }

    function markStageVisited(campaign, stage) {
        mutateEntry(campaign, stage, function (entry) {
            entry.visited = true;
        });
    }

    function markStageSolved(campaign, stage, answer) {
        mutateEntry(campaign, stage, function (entry) {
            entry.solved = true;
            entry.visited = true;
            if (typeof answer === 'string') entry.answer = answer;
        });
    }

    function isStageSolved(campaign, stage) {
        return getStageEntry(campaign, stage).solved === true;
    }

    function markHintRevealed(campaign, stage, index, waitSeconds) {
        var revealAt;
        mutateEntry(campaign, stage, function (entry) {
            var existing = entry.hintsRevealed[index];
            if (typeof existing === 'number') {
                revealAt = existing;
                return;
            }
            revealAt = Date.now() + (waitSeconds > 0 ? waitSeconds * 1000 : 0);
            entry.hintsRevealed[index] = revealAt;
        });
        return revealAt;
    }

    function markCampaignCompleted(campaign) {
        var blob = readBlob();
        if (blob.completedCampaigns.indexOf(campaign) === -1) {
            blob.completedCampaigns.push(campaign);
            writeBlob(blob);
        }
    }

    function isCampaignCompleted(campaign) {
        return readBlob().completedCampaigns.indexOf(campaign) !== -1;
    }

    function getNextUnsolvedStage(campaign, stageOrder) {
        if (!Array.isArray(stageOrder)) return null;
        for (var i = 0; i < stageOrder.length; i++) {
            if (!isStageSolved(campaign, stageOrder[i])) return stageOrder[i];
        }
        return null;
    }

    function computeVisibility(campaign, slug, stageOrder) {
        var entry = getStageEntry(campaign, slug);
        if (entry.visited) return 'visible';
        var next = getNextUnsolvedStage(campaign, stageOrder);
        if (next === slug) return 'next';
        return 'locked';
    }

    async function _sha256(msg) {
        var buf = new TextEncoder().encode(msg);
        var hash = await crypto.subtle.digest('SHA-256', buf);
        return Array.from(new Uint8Array(hash)).map(function (b) {
            return b.toString(16).padStart(2, '0');
        }).join('');
    }

    var currentStage = null;

    function applyPostSolveUI(answer) {
        var input = document.getElementById('tb-answer-input');
        if (input) {
            input.value = answer;
            input.disabled = true;
            input.classList.add('tb-answer-disabled');
        }
        var form = document.querySelector('#answer-box form');
        if (form) {
            var btn = form.querySelector('button[type="submit"]');
            if (btn) btn.style.display = 'none';
        }
        var box = document.getElementById('tb-completion-box');
        if (box) {
            _rw(box);
            box.style.display = 'block';
        }
    }

    function scheduleHintReveal(index, revealAt) {
        var row = document.querySelector('#tb-hint-box-' + index + ' .tb-hint-row');
        var status = document.getElementById('tb-hint-status-' + index);
        var content = document.getElementById('tb-hint-content-' + index);
        if (!row || !status || !content) return;
        if (row.dataset.started) return;
        row.dataset.started = '1';
        row.style.cursor = 'default';

        function reveal() {
            _rn(content.querySelector('[data-z]'));
            row.style.display = 'none';
            content.style.display = 'block';
        }

        if (revealAt - Date.now() <= 0) {
            reveal();
            return;
        }

        var countdown = document.createElement('span');
        countdown.className = 'tb-mono tb-hint-countdown';
        status.replaceChildren(countdown);

        var timer;
        function tick() {
            var leftMs = revealAt - Date.now();
            if (leftMs <= 0) {
                clearInterval(timer);
                reveal();
                return;
            }
            countdown.textContent = Math.ceil(leftMs / 1000) + 's';
        }
        tick();
        timer = setInterval(tick, 1000);
    }

    function revealHint(index, waitSeconds) {
        if (!currentStage) return;
        var row = document.querySelector('#tb-hint-box-' + index + ' .tb-hint-row');
        if (!row || row.dataset.started) return;
        var revealAt = markHintRevealed(currentStage.campaign, currentStage.stage, index, waitSeconds);
        scheduleHintReveal(index, revealAt);
    }

    var ICON_SOLVED = '<svg class="tb-stage-icon tb-stage-icon--solved" viewBox="0 0 20 20" width="12" height="12" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M4 10.5l4 4 8-9"/></svg>';
    var ICON_VISITED = '<svg class="tb-stage-icon tb-stage-icon--visited" viewBox="0 0 20 20" width="12" height="12" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M2.5 10s2.8-5 7.5-5 7.5 5 7.5 5-2.8 5-7.5 5S2.5 10 2.5 10z"/><circle cx="10" cy="10" r="2"/></svg>';

    function setStageIcon(item, html) {
        var host = item.querySelector('.tb-stage-status');
        if (host) {
            host.innerHTML = html;
        } else {
            item.insertAdjacentHTML('afterbegin', html);
        }
    }

    function clearStageIcon(item) {
        var host = item.querySelector('.tb-stage-status');
        if (host) {
            host.innerHTML = '';
            return;
        }
        var existing = item.querySelector(':scope > .tb-stage-icon');
        if (existing) existing.remove();
    }

    function decorateStageItems(items, campaign, stageOrder) {
        if (!items || !Array.isArray(stageOrder)) return;
        items.forEach(function (item) {
            var slug = item.getAttribute('data-tb-stage-slug');
            if (!slug) return;
            var entry = getStageEntry(campaign, slug);
            var vis = computeVisibility(campaign, slug, stageOrder);

            item.classList.remove('tb-stage-locked', 'tb-stage-solved', 'tb-stage-visited', 'tb-stage-next');
            clearStageIcon(item);

            var link = item.tagName === 'A' ? item : item.querySelector('a');

            if (link && link.getAttribute('href') && !link.dataset.tbOriginalHref) {
                link.dataset.tbOriginalHref = link.getAttribute('href');
            }

            if (vis === 'locked') {
                item.classList.add('tb-stage-locked');
                if (link) link.removeAttribute('href');
                return;
            }

            if (link && link.dataset.tbOriginalHref && !link.getAttribute('href')) {
                link.setAttribute('href', link.dataset.tbOriginalHref);
            }

            if (entry.solved) {
                item.classList.add('tb-stage-solved');
                setStageIcon(item, ICON_SOLVED);
            } else if (entry.visited) {
                item.classList.add('tb-stage-visited');
                setStageIcon(item, ICON_VISITED);
            } else if (vis === 'next') {
                item.classList.add('tb-stage-next');
            }
        });
    }

    function decorateStageList(root, campaign, stageOrder) {
        if (!root) return;
        decorateStageItems(root.querySelectorAll('[data-tb-stage-slug]'), campaign, stageOrder);
    }

    function decodePayloads(s) {
        if (!s) return [];
        try {
            var arr = JSON.parse(_zb(s));
            return Array.isArray(arr) ? arr : [];
        } catch (e) {
            return [];
        }
    }

    function renderPips(target, difficulty) {
        if (!target) return;
        var html = '';
        for (var p = 1; p <= 5; p++) {
            html += '<span class="tb-pip' + (p <= difficulty ? ' tb-pip--filled' : '') + '"></span>';
        }
        target.innerHTML = html;
    }

    function revealStagePlaceholders(items, stages, campaign, currentStageSlug) {
        if (!items || !Array.isArray(stages)) return;
        for (var k = 0; k < items.length; k++) {
            var item = items[k];
            var idxAttr = item.getAttribute('data-tb-stage-index');
            if (idxAttr === null) continue;
            var i = parseInt(idxAttr, 10);
            var stage = stages[i];
            if (!stage) continue;

            var prev = i > 0 ? stages[i - 1] : null;
            var entry = getStageEntry(campaign, stage.slug);
            var revealed = (i === 0)
                || (prev && isStageSolved(campaign, prev.slug))
                || (currentStageSlug && stage.slug === currentStageSlug)
                || entry.visited
                || entry.solved;
            if (!revealed) continue;

            item.setAttribute('data-tb-stage-slug', stage.slug);
            item.classList.remove('tb-stage-locked');

            var link = item.tagName === 'A' ? item : item.querySelector('a');
            if (link) {
                link.setAttribute('href', stage.href);
                link.dataset.tbOriginalHref = stage.href;
            }

            var titleEl = item.querySelector('.tb-stage-title');
            if (titleEl) titleEl.textContent = stage.title;

            var redactedTitle = item.querySelector('.tb-stage-title-redacted');
            if (redactedTitle) {
                var parent = redactedTitle.parentNode;
                if (parent) {
                    parent.textContent = stage.title;
                }
            }

            var diffEl = item.querySelector('.tb-stage-difficulty');
            if (diffEl) {
                if (stage.difficulty > 0) {
                    renderPips(diffEl.querySelector('.tb-pips'), stage.difficulty);
                    diffEl.style.display = '';
                } else {
                    diffEl.style.display = 'none';
                }
            }

            var etaEl = item.querySelector('.tb-stage-eta');
            if (etaEl) {
                if (stage.eta > 0) {
                    etaEl.textContent = 'Duration: ' + stage.eta + ' min';
                    etaEl.style.display = '';
                } else {
                    etaEl.style.display = 'none';
                }
            }
        }
    }

    function initStage(opts) {
        if (!opts || !opts.campaign || !opts.stage) return;
        if (typeof opts.k === 'string') TB_K = opts.k;
        currentStage = opts;
        opts._stages = decodePayloads(opts.payloads);
        opts._order = opts._stages.map(function (s) { return s.slug; });

        if (opts.hasAnswer) {
            markStageVisited(opts.campaign, opts.stage);
        } else {
            markStageSolved(opts.campaign, opts.stage, '');
            if (opts.isLast) markCampaignCompleted(opts.campaign);
        }

        var entry = getStageEntry(opts.campaign, opts.stage);
        if (opts.hasAnswer && entry.solved) {
            applyPostSolveUI(entry.answer || '');
        }

        var hints = entry.hintsRevealed || {};
        Object.keys(hints).forEach(function (idx) {
            var revealAt = hints[idx];
            if (typeof revealAt !== 'number') return;
            scheduleHintReveal(parseInt(idx, 10), revealAt);
        });

        refreshStageSidebar();
    }

    function refreshStageSidebar() {
        if (!currentStage) return;
        var sidebar = document.querySelector('.tb-progress-list');
        if (!sidebar) return;
        revealStagePlaceholders(
            sidebar.querySelectorAll('[data-tb-stage-index]'),
            currentStage._stages || [],
            currentStage.campaign,
            currentStage.stage
        );
        decorateStageItems(
            sidebar.querySelectorAll('[data-tb-stage-slug]'),
            currentStage.campaign,
            currentStage._order || []
        );
    }

    async function checkAnswer(event) {
        if (event && typeof event.preventDefault === 'function') event.preventDefault();
        if (!currentStage) return;

        var input = document.getElementById('tb-answer-input');
        var feedback = document.getElementById('tb-answer-feedback');
        if (!input) return;

        var raw = input.value;
        var normalized = raw.trim().toLowerCase();
        if (!normalized) return;

        var hash = await _sha256(normalized);
        if (hash === currentStage.expectedHash) {
            markStageSolved(currentStage.campaign, currentStage.stage, raw);
            if (currentStage.isLast) markCampaignCompleted(currentStage.campaign);
            if (feedback) feedback.innerHTML = '';
            applyPostSolveUI(raw);
            refreshStageSidebar();
        } else if (feedback) {
            feedback.innerHTML = '<div class="tb-notification tb-notification--error">Incorrect — try again.</div>';
        }
    }

    function initCampaign(opts) {
        if (!opts || !opts.campaign) return;
        if (typeof opts.k === 'string') TB_K = opts.k;
        var campaign = opts.campaign;
        var stages = decodePayloads(opts.payloads);
        var order = stages.map(function (s) { return s.slug; });
        var titles = {};
        var hrefs = {};
        stages.forEach(function (s) { titles[s.slug] = s.title; hrefs[s.slug] = s.href; });
        var totalStages = order.length;

        var placeholders = document.querySelectorAll('[data-tb-stage-index]');
        revealStagePlaceholders(placeholders, stages, campaign, null);

        var solvedCount = 0;
        var visitedCount = 0;
        for (var i = 0; i < order.length; i++) {
            var e = getStageEntry(campaign, order[i]);
            if (e.solved) solvedCount++;
            if (e.solved || e.visited) visitedCount++;
        }
        var completed = isCampaignCompleted(campaign);
        var hasProgress = visitedCount > 0 || completed;

        var next = getNextUnsolvedStage(campaign, order);
        var nextHref = next ? hrefs[next] : null;

        if (hasProgress) {
            var bar = document.getElementById('tb-campaign-progress');
            var label = document.getElementById('tb-campaign-progress-label');
            var fill = document.getElementById('tb-campaign-progress-fill');
            if (bar) bar.style.display = 'block';
            if (label) label.textContent = solvedCount + ' of ' + totalStages + ' stages cleared';
            if (fill) {
                var pct = totalStages > 0 ? (solvedCount / totalStages) * 100 : 0;
                fill.style.width = pct + '%';
            }
        }

        if (completed) {
            var tag = document.getElementById('tb-campaign-completed-tag');
            if (tag) tag.style.display = '';
            var section = document.getElementById('tb-campaign-completion-section');
            if (section) {
                _rw(section);
                section.style.display = '';
            }
        }

        var action = document.getElementById('tb-campaign-action');
        if (action) {
            if (completed) {
                action.innerHTML = '<div class="tb-btn tb-btn-primary" style="width: 100%; justify-content: center; cursor: default;">✓ Campaign completed</div>';
            } else if (visitedCount > 0 && next && nextHref) {
                var nextTitle = titles[next] || next;
                var safeTitle = nextTitle.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
                var safeHref = nextHref.replace(/"/g, '&quot;');
                action.innerHTML = '<a class="tb-btn tb-btn-primary tb-btn-glow" href="' + safeHref + '" style="width: 100%; justify-content: center;">Continue from ' + safeTitle + ' →</a>';
            }
        }

        decorateStageItems(document.querySelectorAll('[data-tb-stage-slug]'), campaign, order);
    }

    function hasCampaignProgress(campaign, stageOrder) {
        if (!Array.isArray(stageOrder)) return false;
        for (var i = 0; i < stageOrder.length; i++) {
            if (getStageEntry(campaign, stageOrder[i]).solved) return true;
        }
        return false;
    }

    function addRibbon(card, kind) {
        if (!card) return;
        if (card.querySelector('.tb-ribbon')) return;
        var label = kind === 'completed' ? 'Completed' : 'In progress';
        var cls = kind === 'completed' ? 'tb-ribbon--completed' : 'tb-ribbon--in-progress';
        var span = document.createElement('span');
        span.className = 'tb-ribbon ' + cls;
        span.textContent = label;
        card.appendChild(span);
    }

    function makeCompletedRow(card) {
        var href = card.getAttribute('href') || '#';
        var titleEl = card.querySelector('.tb-card-title');
        var title = titleEl ? titleEl.textContent.trim() : (card.getAttribute('data-tb-campaign-id') || '');

        var row = document.createElement('a');
        row.className = 'tb-recessed-row';
        row.href = href;

        var icon = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
        icon.setAttribute('class', 'tb-recessed-row-icon');
        icon.setAttribute('viewBox', '0 0 20 20');
        icon.setAttribute('width', '14');
        icon.setAttribute('height', '14');
        icon.setAttribute('fill', 'none');
        icon.setAttribute('stroke', 'currentColor');
        icon.setAttribute('stroke-width', '2');
        icon.setAttribute('stroke-linecap', 'round');
        icon.setAttribute('stroke-linejoin', 'round');
        icon.setAttribute('aria-hidden', 'true');
        var path = document.createElementNS('http://www.w3.org/2000/svg', 'path');
        path.setAttribute('d', 'M4 10.5l4 4 8-9');
        icon.appendChild(path);

        var titleSpan = document.createElement('span');
        titleSpan.className = 'tb-recessed-row-title';
        titleSpan.textContent = title;

        var chipsWrap = document.createElement('div');
        chipsWrap.className = 'tb-recessed-row-chips';
        var srcChips = card.querySelectorAll('.tb-card-body .tb-chip');
        srcChips.forEach(function (chip) {
            chipsWrap.appendChild(chip.cloneNode(true));
        });

        var arrow = document.createElement('span');
        arrow.className = 'tb-recessed-row-arrow';
        arrow.setAttribute('aria-hidden', 'true');
        arrow.textContent = '→';

        row.appendChild(icon);
        row.appendChild(titleSpan);
        if (chipsWrap.children.length > 0) row.appendChild(chipsWrap);
        row.appendChild(arrow);
        return row;
    }

    function initHome(opts) {
        if (!opts) return;
        var featured = Array.isArray(opts.featured) ? opts.featured : [];
        var nonFeatured = Array.isArray(opts.nonFeatured) ? opts.nonFeatured : [];

        var availableGrid = document.getElementById('tb-available-grid');
        var availableSection = document.getElementById('tb-available');
        var completedSection = document.getElementById('tb-completed');
        var completedList = document.getElementById('tb-completed-list');

        featured.forEach(function (c) {
            if (!c || !c.id) return;
            var card = document.querySelector('.tb-card--featured[data-tb-campaign-id="' + c.id + '"]');
            if (!card) return;
            if (isCampaignCompleted(c.id)) {
                addRibbon(card, 'completed');
                if (completedList) {
                    completedList.appendChild(makeCompletedRow(card));
                }
            } else if (hasCampaignProgress(c.id, c.order)) {
                addRibbon(card, 'in-progress');
            }
        });

        nonFeatured.forEach(function (c) {
            if (!c || !c.id) return;
            var card = document.querySelector('.tb-card--available[data-tb-campaign-id="' + c.id + '"]');
            if (!card) return;
            if (isCampaignCompleted(c.id)) {
                if (completedList) {
                    completedList.appendChild(makeCompletedRow(card));
                }
                var column = card.closest('.column');
                if (column && column.parentNode) {
                    column.parentNode.removeChild(column);
                } else if (card.parentNode) {
                    card.parentNode.removeChild(card);
                }
            } else if (hasCampaignProgress(c.id, c.order)) {
                addRibbon(card, 'in-progress');
            }
        });

        if (availableGrid && availableSection) {
            if (!availableGrid.querySelector('.tb-card--available')) {
                availableSection.style.display = 'none';
            }
        }
        if (completedList && completedSection) {
            if (completedList.children.length > 0) {
                completedSection.style.display = '';
            }
        }

        renumberSections();
    }

    function renumberSections() {
        var sections = document.querySelectorAll('.tb-block');
        var n = 0;
        sections.forEach(function (sec) {
            if (window.getComputedStyle(sec).display === 'none') return;
            var label = sec.querySelector('.tb-section-num');
            if (!label) return;
            n += 1;
            label.textContent = (n < 10 ? '0' : '') + n;
        });
    }

    function normalizePath(p) {
        p = (p || '').split('#')[0].split('?')[0];
        p = p.replace(/\/index\.html?$/i, '/');
        if (p.length > 1) p = p.replace(/\/+$/, '');
        return p || '/';
    }

    function markActiveNav() {
        var links = document.querySelectorAll('.tb-nav-menu a');
        var current = normalizePath(window.location.pathname);
        links.forEach(function (a) {
            var href = a.getAttribute('href');
            if (!href || /^[a-z]+:/i.test(href)) return; // skip external/absolute
            if (normalizePath(href) === current) a.classList.add('is-active');
        });
    }

    /* btoa/atob only handle Latin-1; stored answers keep raw user input, so
       round-trip through UTF-8 bytes to survive accents and emoji. */
    function encodeBase64Utf8(str) {
        var bytes = new TextEncoder().encode(str);
        var bin = '';
        for (var i = 0; i < bytes.length; i++) bin += String.fromCharCode(bytes[i]);
        return btoa(bin);
    }

    function decodeBase64Utf8(b64) {
        var bin = atob(b64);
        var bytes = new Uint8Array(bin.length);
        for (var i = 0; i < bin.length; i++) bytes[i] = bin.charCodeAt(i);
        return new TextDecoder().decode(bytes);
    }

    function exportProgress() {
        var raw = window.localStorage.getItem(STORAGE_KEY);
        if (!raw) raw = JSON.stringify(defaultBlob());
        return encodeBase64Utf8(raw);
    }

    function importProgress(code) {
        var trimmed = (code || '').trim();
        if (!trimmed) return 'Paste a progress code first.';
        var json;
        try {
            json = decodeBase64Utf8(trimmed);
        } catch (e) {
            return 'Invalid progress code.';
        }
        var parsed;
        try {
            parsed = JSON.parse(json);
        } catch (e) {
            return 'Invalid progress code.';
        }
        if (!parsed || typeof parsed !== 'object') return 'Invalid progress code.';
        if (parsed.version !== SCHEMA_VERSION) return 'Progress code is from an incompatible version.';
        if (!Array.isArray(parsed.completedCampaigns)) parsed.completedCampaigns = [];
        if (!parsed.stageProgress || typeof parsed.stageProgress !== 'object') parsed.stageProgress = {};
        writeBlob(parsed);
        return null;
    }

    function resetProgress() {
        try {
            window.localStorage.removeItem(STORAGE_KEY);
        } catch (e) {
            /* disabled storage — nothing to clear */
        }
    }

    function initProgressModal() {
        var modal = document.getElementById('tb-progress-modal');
        var trigger = document.getElementById('tb-progress-open');
        if (!modal || !trigger) return;

        var exportField = document.getElementById('tb-progress-export');
        var importField = document.getElementById('tb-progress-import');
        var copyBtn = document.getElementById('tb-progress-copy');
        var importBtn = document.getElementById('tb-progress-import-btn');
        var resetBtn = document.getElementById('tb-progress-reset');
        var feedback = document.getElementById('tb-progress-feedback');

        function showFeedback(kind, msg) {
            if (!feedback) return;
            feedback.innerHTML = '<div class="tb-notification tb-notification--' + kind + '">' + msg + '</div>';
        }

        function clearFeedback() {
            if (feedback) feedback.innerHTML = '';
        }

        function open() {
            clearFeedback();
            if (importField) importField.value = '';
            if (exportField) exportField.value = exportProgress();
            modal.hidden = false;
            document.body.classList.add('tb-modal-open');
        }

        function close() {
            modal.hidden = true;
            document.body.classList.remove('tb-modal-open');
        }

        trigger.addEventListener('click', open);

        modal.querySelectorAll('[data-tb-progress-close]').forEach(function (el) {
            el.addEventListener('click', close);
        });

        document.addEventListener('keydown', function (e) {
            if (e.key === 'Escape' && !modal.hidden) close();
        });

        if (copyBtn && exportField) {
            copyBtn.addEventListener('click', function () {
                var value = exportField.value;
                function done() {
                    var prev = copyBtn.textContent;
                    copyBtn.textContent = 'Copied';
                    setTimeout(function () { copyBtn.textContent = prev; }, 1500);
                }
                if (navigator.clipboard && navigator.clipboard.writeText) {
                    navigator.clipboard.writeText(value).then(done, function () {
                        exportField.select();
                        try { document.execCommand('copy'); } catch (e) {}
                        done();
                    });
                } else {
                    exportField.select();
                    try { document.execCommand('copy'); } catch (e) {}
                    done();
                }
            });
        }

        if (importBtn) {
            importBtn.addEventListener('click', function () {
                var err = importProgress(importField ? importField.value : '');
                if (err) {
                    showFeedback('error', err);
                    return;
                }
                showFeedback('success', 'Progress imported. Reloading…');
                window.location.reload();
            });
        }

        if (resetBtn) {
            resetBtn.addEventListener('click', function () {
                if (!window.confirm('Reset all progress on this browser? This cannot be undone.')) return;
                resetProgress();
                showFeedback('success', 'Progress reset. Reloading…');
                window.location.reload();
            });
        }
    }

    function initChrome() {
        markActiveNav();
        initProgressModal();
    }

    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', initChrome);
    } else {
        initChrome();
    }

    window.TBProgress = {
        initHome: initHome,
        initCampaign: initCampaign,
        initStage: initStage,
        checkAnswer: checkAnswer,
        revealHint: revealHint,
        exportProgress: exportProgress,
        importProgress: importProgress,
        resetProgress: resetProgress
    };
})();
