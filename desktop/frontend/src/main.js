import './style.css';
import './app.css';

// Import Wails runtime
import { GetHealth, GetSkills, GenerateCode, ReviewCode, DeployPlugin, MCPListDocs, MCPGetDoc, MCPSearchDocs, MCPGetHookInventory, RuntimeStart, RuntimeStop, RuntimeGetStatus, RuntimeGetOutput } from '../wailsjs/go/main/App';

let currentLang = 'en';
let ws = null;
let wsEvents = [];

// i18n translations
const i18n = {
  en: {
    newConversation: 'New conversation',
    modules: 'MODULES',
    chatModule: 'Chat',
    generateModule: 'Generate',
    reviewModule: 'Review',
    deployModule: 'Deploy',
    skillsModule: 'Skills',
    mcpModule: 'MCP Tools',
    runtimeModule: 'Runtime',
    system: 'SYSTEM',
    healthModule: 'Health',
    wsModule: 'Events',
    welcomeTitle: 'Welcome to AstrCode',
    connecting: 'Connecting...',
    connected: 'Connected',
    disconnected: 'Disconnected',
    welcomeMessage: "Hello! I'm your AI coding assistant. How can I help you today?",
    chatPlaceholder: 'Ask anything about your codebase...',
    send: 'Send',
    generateDescription: 'Describe what you want to create, and I\'ll help you generate the code.',
    generatePlaceholder: 'e.g., Create a weather plugin that supports city name queries...',
    reviewDescription: 'Paste your code here, and I\'ll provide a comprehensive review.',
    reviewPlaceholder: 'Paste your code here...',
    deployDescription: 'Ready to deploy your plugin? Provide the details and I\'ll help you.',
    deployPlaceholder: 'Plugin name, description, files...',
    skillsTitle: 'Skills',
    skillsDescription: 'Available skills and plugins in your AstrCode installation.',
    skillsClickToLoad: 'Click to load skills',
    mcpTitle: 'MCP Tools',
    mcpDescription: 'AstrBot-Skill MCP Server - Browse documentation, search docs, and view hook inventory.',
    runtimeTitle: 'AstrBot Runtime',
    runtimeDescription: 'Start and monitor AstrBot runtime with WebUI and terminal output.',
    healthTitle: 'System Health',
    healthDescription: 'Check the status of your AstrCode server.',
    healthClickToCheck: 'Click to check health status',
    wsTitle: 'WebSocket Events',
    wsDescription: 'Real-time event stream from the server.',
    wsStatus: 'WebSocket connection status',
    inputFooter: 'AstrCode can make mistakes. Check important info.',
    user: 'You',
    assistant: 'AstrCode'
  },
  zh: {
    newConversation: '新对话',
    modules: '功能模块',
    chatModule: '对话',
    generateModule: '生成',
    reviewModule: '审查',
    deployModule: '部署',
    skillsModule: '技能',
    mcpModule: 'MCP 工具',
    runtimeModule: '运行时',
    system: '系统',
    healthModule: '健康',
    wsModule: '事件',
    welcomeTitle: '欢迎使用 AstrCode',
    connecting: '连接中...',
    connected: '已连接',
    disconnected: '未连接',
    welcomeMessage: '你好！我是你的 AI 编程助手。今天我能帮你什么？',
    chatPlaceholder: '询问关于你的代码库的任何问题...',
    send: '发送',
    generateDescription: '描述你想要创建的内容，我会帮你生成代码。',
    generatePlaceholder: '例如：创建一个支持城市名称查询的天气插件...',
    reviewDescription: '在这里粘贴你的代码，我会提供全面的审查意见。',
    reviewPlaceholder: '在这里粘贴你的代码...',
    deployDescription: '准备部署你的插件？提供详细信息，我会帮助你。',
    deployPlaceholder: '插件名称、描述、文件...',
    skillsTitle: '技能',
    skillsDescription: 'AstrCode 安装中可用的技能和插件。',
    skillsClickToLoad: '点击加载技能列表',
    mcpTitle: 'MCP 工具',
    mcpDescription: 'AstrBot-Skill MCP 服务器 - 浏览文档、搜索文档和查看 Hook 清单。',
    runtimeTitle: 'AstrBot 运行时',
    runtimeDescription: '启动并监控 AstrBot 运行时，包含 WebUI 和终端输出。',
    healthTitle: '系统健康',
    healthDescription: '检查 AstrCode 服务器的状态。',
    healthClickToCheck: '点击检查健康状态',
    wsTitle: 'WebSocket 事件',
    wsDescription: '来自服务器的实时事件流。',
    wsStatus: 'WebSocket 连接状态',
    inputFooter: 'AstrCode 可能会出错。请检查重要信息。',
    user: '你',
    assistant: 'AstrCode'
  }
};

function t(key) {
  return i18n[currentLang][key] || i18n.en[key] || key;
}

function toggleLanguage() {
  currentLang = currentLang === 'en' ? 'zh' : 'en';
  document.getElementById('current-lang').textContent = currentLang === 'en' ? 'English' : '中文';
  document.documentElement.lang = currentLang === 'en' ? 'en' : 'zh-CN';
  updateTranslations();
}

function updateTranslations() {
  document.querySelectorAll('[data-i18n]').forEach(el => {
    const key = el.getAttribute('data-i18n');
    el.textContent = t(key);
  });
  document.querySelectorAll('[data-i18n-placeholder]').forEach(el => {
    const key = el.getAttribute('data-i18n-placeholder');
    el.placeholder = t(key);
  });
}

function switchModule(moduleName) {
  document.querySelectorAll('.module').forEach(m => m.classList.remove('active'));
  document.querySelectorAll('.nav-item').forEach(n => n.classList.remove('active'));
  
  document.getElementById(`module-${moduleName}`).classList.add('active');
  document.querySelector(`[data-module="${moduleName}"]`).classList.add('active');

  // Auto-load data for specific modules
  if (moduleName === 'health') {
    setTimeout(() => loadHealth(), 100);
  } else if (moduleName === 'skills') {
    setTimeout(() => loadSkills(), 100);
  } else if (moduleName === 'mcp') {
    setTimeout(() => initMCPModule(), 100);
  } else if (moduleName === 'runtime') {
    setTimeout(() => initRuntimeModule(), 100);
  }
}

function addMessage(module, role, content, isStreaming = false) {
  const container = document.getElementById(`${module}-messages`);
  const messageDiv = document.createElement('div');
  messageDiv.className = `message ${role}`;
  
  const time = new Date().toLocaleTimeString(currentLang === 'zh' ? 'zh-CN' : 'en', { 
    hour: '2-digit', 
    minute: '2-digit' 
  });
  
  messageDiv.innerHTML = `
    <div class="message-avatar ${role}">${role === 'user' ? 'U' : 'A'}</div>
    <div class="message-content">
      <div class="message-header">
        <span class="message-name">${role === 'user' ? t('user') : t('assistant')}</span>
        <span class="message-time">${time}</span>
      </div>
      <div class="message-body">${content}</div>
      ${isStreaming ? '<div class="streaming-indicator"><div class="streaming-dot"></div><div class="streaming-dot"></div><div class="streaming-dot"></div></div>' : ''}
    </div>
  `;
  container.appendChild(messageDiv);
  messageDiv.scrollIntoView({ behavior: 'smooth' });
  return messageDiv;
}

function updateMessage(messageDiv, content) {
  messageDiv.querySelector('.message-body').innerHTML = content;
}

function newConversation() {
  location.reload();
}

// Auto-resize textarea
document.querySelectorAll('.chat-input').forEach(textarea => {
  textarea.addEventListener('input', function() {
    this.style.height = 'auto';
    this.style.height = Math.min(this.scrollHeight, 200) + 'px';
  });
  
  textarea.addEventListener('keydown', function(e) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      const module = this.id.split('-')[0];
      const sendBtn = document.getElementById(`${module}-send`);
      if (sendBtn && !sendBtn.disabled) {
        sendBtn.click();
      }
    }
  });
});

// WebSocket
// WebSocket connection disabled (using Wails native API)
// function connectWS() { ... }

// API calls with streaming simulation
async function sendChat() {
  const input = document.getElementById('chat-input');
  const sendBtn = document.getElementById('chat-send');
  const message = input.value.trim();
  if (!message) return;

  addMessage('chat', 'user', message);
  input.value = '';
  input.style.height = 'auto';
  sendBtn.disabled = true;

  const assistantMsg = addMessage('chat', 'assistant', '', true);

  try {
    const data = await GenerateCode(message);
    
    // Simulate streaming effect
    let displayText = '';
    const jsonStr = JSON.stringify(JSON.parse(data), null, 2);
    const chunks = jsonStr.split('');
    
    for (let i = 0; i < chunks.length; i++) {
      displayText += chunks[i];
      updateMessage(assistantMsg, `<pre><code>${displayText}</code></pre>`);
      await new Promise(resolve => setTimeout(resolve, 5));
    }
    
    // Remove streaming indicator
    const indicator = assistantMsg.querySelector('.streaming-indicator');
    if (indicator) indicator.remove();
  } catch (error) {
    updateMessage(assistantMsg, `<p style="color: var(--error);">Error: ${error.message}</p>`);
    const indicator = assistantMsg.querySelector('.streaming-indicator');
    if (indicator) indicator.remove();
  } finally {
    sendBtn.disabled = false;
  }
}

async function generateCode() {
  const input = document.getElementById('generate-input');
  const sendBtn = document.getElementById('generate-send');
  const requirement = input.value.trim();
  if (!requirement) return;

  addMessage('generate', 'user', requirement);
  input.value = '';
  input.style.height = 'auto';
  sendBtn.disabled = true;

  const assistantMsg = addMessage('generate', 'assistant', '', true);

  try {
    const data = await GenerateCode(requirement);
    
    let displayText = '';
    const jsonStr = JSON.stringify(JSON.parse(data), null, 2);
    const chunks = jsonStr.split('');
    
    for (let i = 0; i < chunks.length; i++) {
      displayText += chunks[i];
      updateMessage(assistantMsg, `<pre><code>${displayText}</code></pre>`);
      await new Promise(resolve => setTimeout(resolve, 5));
    }
    
    const indicator = assistantMsg.querySelector('.streaming-indicator');
    if (indicator) indicator.remove();
  } catch (error) {
    updateMessage(assistantMsg, `<p style="color: var(--error);">Error: ${error.message}</p>`);
    const indicator = assistantMsg.querySelector('.streaming-indicator');
    if (indicator) indicator.remove();
  } finally {
    sendBtn.disabled = false;
  }
}

async function reviewCode() {
  const input = document.getElementById('review-input');
  const sendBtn = document.getElementById('review-send');
  const code = input.value.trim();
  if (!code) return;

  addMessage('review', 'user', `<pre><code>${code}</code></pre>`);
  input.value = '';
  input.style.height = 'auto';
  sendBtn.disabled = true;

  const assistantMsg = addMessage('review', 'assistant', '', true);

  try {
    const data = await ReviewCode(code);
    
    let displayText = '';
    const jsonStr = JSON.stringify(JSON.parse(data), null, 2);
    const chunks = jsonStr.split('');
    
    for (let i = 0; i < chunks.length; i++) {
      displayText += chunks[i];
      updateMessage(assistantMsg, `<pre><code>${displayText}</code></pre>`);
      await new Promise(resolve => setTimeout(resolve, 5));
    }
    
    const indicator = assistantMsg.querySelector('.streaming-indicator');
    if (indicator) indicator.remove();
  } catch (error) {
    updateMessage(assistantMsg, `<p style="color: var(--error);">Error: ${error.message}</p>`);
    const indicator = assistantMsg.querySelector('.streaming-indicator');
    if (indicator) indicator.remove();
  } finally {
    sendBtn.disabled = false;
  }
}

async function deployPlugin() {
  const input = document.getElementById('deploy-input');
  const sendBtn = document.getElementById('deploy-send');
  const deployInfo = input.value.trim();
  if (!deployInfo) return;

  addMessage('deploy', 'user', deployInfo);
  input.value = '';
  input.style.height = 'auto';
  sendBtn.disabled = true;

  const assistantMsg = addMessage('deploy', 'assistant', '', true);

  try {
    const data = await DeployPlugin('my-plugin', deployInfo);
    
    let displayText = '';
    const jsonStr = JSON.stringify(JSON.parse(data), null, 2);
    const chunks = jsonStr.split('');
    
    for (let i = 0; i < chunks.length; i++) {
      displayText += chunks[i];
      updateMessage(assistantMsg, `<pre><code>${displayText}</code></pre>`);
      await new Promise(resolve => setTimeout(resolve, 5));
    }
    
    const indicator = assistantMsg.querySelector('.streaming-indicator');
    if (indicator) indicator.remove();
  } catch (error) {
    updateMessage(assistantMsg, `<p style="color: var(--error);">Error: ${error.message}</p>`);
    const indicator = assistantMsg.querySelector('.streaming-indicator');
    if (indicator) indicator.remove();
  } finally {
    sendBtn.disabled = false;
  }
}

async function loadSkills() {
  try {
    const data = await GetSkills();
    const skillsData = JSON.parse(data);
    const skillsList = document.getElementById('skills-list');
    
    if (!skillsData.skills || skillsData.skills.length === 0) {
      skillsList.innerHTML = `
        <div class="empty-state" style="grid-column: 1 / -1;">
          <svg class="empty-state-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
            <path d="M12 2L2 7l10 5 10-5-10-5z"/>
            <path d="M2 17l10 5 10-5"/>
            <path d="M2 12l10 5 10-5"/>
          </svg>
          <div class="empty-state-text">No skills installed</div>
          <div class="empty-state-hint">Install skills to enhance your coding experience</div>
        </div>
      `;
      return;
    }
    
    skillsList.innerHTML = skillsData.skills.map(skill => `
      <div class="skill-card">
        <div class="skill-header">
          <div class="skill-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/>
            </svg>
          </div>
          <div>
            <h3 class="skill-title">${skill.name || 'Unknown Skill'}</h3>
            <div class="skill-version">v${skill.version || '1.0.0'}</div>
          </div>
        </div>
        <div class="skill-description">${skill.description || 'No description available'}</div>
        <div class="skill-tags">
          ${skill.tags ? skill.tags.map(tag => `<span class="skill-tag">${tag}</span>`).join('') : ''}
        </div>
      </div>
    `).join('');
  } catch (error) {
    document.getElementById('skills-list').innerHTML = `<p style="color: var(--error);">Error: ${error.message}</p>`;
  }
}

async function loadHealth() {
  try {
    const data = await GetHealth();
    const healthData = JSON.parse(data);
    const healthResult = document.getElementById('health-result');
    
    const isHealthy = healthData.status === 'ok';
    
    healthResult.innerHTML = `
      <div class="health-status-card ${isHealthy ? 'healthy' : 'unhealthy'}">
        <div class="health-status-icon">
          ${isHealthy ? `
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/>
              <polyline points="22 4 12 14.01 9 11.01"/>
            </svg>
          ` : `
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="12" cy="12" r="10"/>
              <line x1="15" y1="9" x2="9" y2="15"/>
              <line x1="9" y1="9" x2="15" y2="15"/>
            </svg>
          `}
        </div>
        <div class="health-status-text">${isHealthy ? 'System Healthy' : 'System Unhealthy'}</div>
        <div class="health-status-subtitle">Status: ${healthData.status || 'unknown'}</div>
      </div>
      
      <div class="health-metrics">
        <div class="health-metric">
          <div class="health-metric-label">Version</div>
          <div class="health-metric-value">${healthData.version || 'N/A'}</div>
        </div>
        <div class="health-metric">
          <div class="health-metric-label">Skills</div>
          <div class="health-metric-value">${healthData.skills_loaded || 0}</div>
        </div>
        <div class="health-metric">
          <div class="health-metric-label">MCP Tools</div>
          <div class="health-metric-value">${healthData.mcp_tools || 0}</div>
        </div>
        <div class="health-metric">
          <div class="health-metric-label">Docs</div>
          <div class="health-metric-value">${healthData.docs_available || 0}</div>
        </div>
        <div class="health-metric">
          <div class="health-metric-label">Hooks</div>
          <div class="health-metric-value">${healthData.hooks_documented || 0}</div>
        </div>
        <div class="health-metric">
          <div class="health-metric-label">Status</div>
          <div class="health-metric-value" style="color: ${isHealthy ? 'var(--success)' : 'var(--error)'}; font-size: 18px;">
            ${isHealthy ? '✓ OK' : '✗ Error'}
          </div>
        </div>
      </div>
    `;
  } catch (error) {
    document.getElementById('health-result').innerHTML = `<p style="color: var(--error);">Error: ${error.message}</p>`;
  }
}

// MCP Module Functions
let selectedMCPTool = 'list_docs';

function initMCPModule() {
  // Bind MCP tool card clicks
  document.querySelectorAll('.mcp-tool-card').forEach(card => {
    card.addEventListener('click', () => {
      document.querySelectorAll('.mcp-tool-card').forEach(c => c.classList.remove('active'));
      card.classList.add('active');
      selectedMCPTool = card.getAttribute('data-mcp-tool');
      updateMCPInputVisibility();
    });
  });
  
  // Select first tool by default
  const firstTool = document.querySelector('.mcp-tool-card');
  if (firstTool) {
    firstTool.classList.add('active');
    selectedMCPTool = 'list_docs';
  }
  
  updateMCPInputVisibility();
}

function updateMCPInputVisibility() {
  const categorySelect = document.getElementById('mcp-category-select');
  const docInput = document.getElementById('mcp-doc-input');
  const searchInput = document.getElementById('mcp-search-input');
  
  // Reset visibility
  categorySelect.style.display = 'block';
  docInput.style.display = 'none';
  searchInput.style.display = 'none';
  
  switch (selectedMCPTool) {
    case 'list_docs':
      // Category select is visible by default
      break;
    case 'get_doc':
      docInput.style.display = 'block';
      break;
    case 'search_docs':
      searchInput.style.display = 'block';
      categorySelect.style.display = 'none';
      break;
    case 'get_hook_inventory':
      categorySelect.style.display = 'none';
      break;
  }
}

async function executeMCPTool() {
  const resultDiv = document.getElementById('mcp-result');
  const category = document.getElementById('mcp-category-select').value;
  const docName = document.getElementById('mcp-doc-input').value.trim();
  const searchQuery = document.getElementById('mcp-search-input').value.trim();
  
  resultDiv.innerHTML = `
    <div class="mcp-loading">
      <div class="streaming-indicator">
        <div class="streaming-dot"></div>
        <div class="streaming-dot"></div>
        <div class="streaming-dot"></div>
      </div>
      <span>Executing ${selectedMCPTool}...</span>
    </div>
  `;
  
  try {
    let result;
    switch (selectedMCPTool) {
      case 'list_docs':
        result = await MCPListDocs(category);
        break;
      case 'get_doc':
        if (!docName) {
          resultDiv.innerHTML = `<p style="color: var(--error);">Please enter a document name</p>`;
          return;
        }
        result = await MCPGetDoc(category || 'plugin_config', docName);
        break;
      case 'search_docs':
        if (!searchQuery) {
          resultDiv.innerHTML = `<p style="color: var(--error);">Please enter a search keyword</p>`;
          return;
        }
        result = await MCPSearchDocs(searchQuery);
        break;
      case 'get_hook_inventory':
        result = await MCPGetHookInventory();
        break;
      default:
        result = `{"error":"Unknown tool: ${selectedMCPTool}"}`;
    }
    
    // Try to parse and format as JSON, fallback to raw text
    let formattedContent;
    try {
      const parsed = JSON.parse(result);
      if (parsed.content) {
        formattedContent = formatMarkdown(parsed.content);
      } else if (parsed.results) {
        formattedContent = `<pre><code>${JSON.stringify(parsed, null, 2)}</code></pre>`;
      } else {
        formattedContent = `<pre><code>${JSON.stringify(parsed, null, 2)}</code></pre>`;
      }
    } catch (e) {
      formattedContent = `<pre><code>${result}</code></pre>`;
    }
    
    resultDiv.innerHTML = `
      <div class="mcp-result-item">
        <div class="mcp-result-tool">${selectedMCPTool}</div>
        <div class="mcp-result-body">${formattedContent}</div>
      </div>
    `;
  } catch (error) {
    resultDiv.innerHTML = `<p style="color: var(--error);">Error: ${error.message}</p>`;
  }
}

function formatMarkdown(text) {
  // Simple markdown formatting
  return text
    .replace(/^# (.*$)/gim, '<h1>$1</h1>')
    .replace(/^## (.*$)/gim, '<h2>$2</h2>')
    .replace(/^### (.*$)/gim, '<h3>$3</h3>')
    .replace(/`([^`]+)`/g, '<code>$1</code>')
    .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
    .replace(/^- (.*$)/gim, '<li>$1</li>')
    .replace(/(<li>.*<\/li>)/s, '<ul>$1</ul>')
    .replace(/\n/g, '<br>');
}

function clearMCPResult() {
  document.getElementById('mcp-result').innerHTML = `
    <div class="mcp-welcome">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M12 2L2 7l10 5 10-5-10-5z"></path><path d="M2 17l10 5 10-5"></path><path d="M2 12l10 5 10-5"></path></svg>
      <p>Select a tool above and click Execute to get started</p>
    </div>
  `;
}

// Initialize
document.addEventListener('DOMContentLoaded', () => {
    // Bind navigation clicks
    document.querySelectorAll('.nav-item').forEach(item => {
        item.addEventListener('click', () => {
            const module = item.getAttribute('data-module');
            switchModule(module);
        });
    });

    // Bind send buttons
    document.getElementById('chat-send')?.addEventListener('click', sendChat);
    document.getElementById('generate-send')?.addEventListener('click', generateCode);
    document.getElementById('review-send')?.addEventListener('click', reviewCode);
    document.getElementById('deploy-send')?.addEventListener('click', deployPlugin);

    // Bind language toggle
    document.querySelector('.language-selector')?.addEventListener('click', toggleLanguage);

    // Bind new conversation
    document.querySelector('.new-conversation-btn')?.addEventListener('click', newConversation);

    // Bind MCP buttons
    document.getElementById('mcp-execute-btn')?.addEventListener('click', executeMCPTool);
    document.getElementById('mcp-clear-btn')?.addEventListener('click', clearMCPResult);

    // Bind skills card click - disabled, auto-load on module switch
    // document.querySelector('#skills-list .api-card')?.addEventListener('click', loadSkills);

    // Bind health card click - disabled, auto-load on module switch
    // document.querySelector('#module-health .api-card')?.addEventListener('click', loadHealth);

    // Auto-resize textarea
    document.querySelectorAll('.chat-input').forEach(textarea => {
        textarea.addEventListener('input', function() {
            this.style.height = 'auto';
            this.style.height = Math.min(this.scrollHeight, 200) + 'px';
        });
        
        textarea.addEventListener('keydown', function(e) {
            if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                const module = this.id.split('-')[0];
                const sendBtn = document.getElementById(`${module}-send`);
                if (sendBtn && !sendBtn.disabled) {
                    sendBtn.click();
                }
            }
        });
    });

    // WebSocket disabled - using Wails native API
    // connectWS();
    updateTranslations();
});

// Runtime Module Functions
let runtimePollingInterval = null;

async function initRuntimeModule() {
  const startBtn = document.getElementById('runtime-start-btn');
  const stopBtn = document.getElementById('runtime-stop-btn');
  const clearBtn = document.getElementById('clear-terminal-btn');
  
  if (startBtn) {
    startBtn.addEventListener('click', startRuntime);
  }
  
  if (stopBtn) {
    stopBtn.addEventListener('click', stopRuntime);
  }
  
  if (clearBtn) {
    clearBtn.addEventListener('click', clearTerminal);
  }
  
  // Check initial status
  await checkRuntimeStatus();
}

async function startRuntime() {
  const pathInput = document.getElementById('astrbot-path');
  const astrbotPath = pathInput.value.trim();
  
  if (!astrbotPath) {
    appendTerminalLine('Error: Please provide AstrBot path', 'error');
    return;
  }
  
  appendTerminalLine(`Starting AstrBot from: ${astrbotPath}...`, 'info');
  
  try {
    const result = await RuntimeStart(astrbotPath);
    const data = JSON.parse(result);
    
    if (data.success) {
      appendTerminalLine('✓ AstrBot runtime started successfully', 'success');
      appendTerminalLine('Waiting for WebUI to become available...', 'info');
      
      // Update UI
      document.getElementById('runtime-start-btn').disabled = true;
      document.getElementById('runtime-stop-btn').disabled = false;
      document.getElementById('runtime-status-dot').classList.add('running');
      document.getElementById('runtime-status-text').textContent = 'Running';
      
      // Start polling for output and status
      startRuntimePolling();
      
      // Wait a bit then load WebUI
      setTimeout(() => {
        loadWebUI();
      }, 3000);
    } else {
      appendTerminalLine(`✗ Failed to start: ${data.error}`, 'error');
    }
  } catch (error) {
    appendTerminalLine(`✗ Error: ${error.message}`, 'error');
  }
}

async function stopRuntime() {
  appendTerminalLine('Stopping AstrBot runtime...', 'info');
  
  try {
    const result = await RuntimeStop();
    const data = JSON.parse(result);
    
    if (data.success) {
      appendTerminalLine('✓ AstrBot runtime stopped', 'success');
      
      // Update UI
      document.getElementById('runtime-start-btn').disabled = false;
      document.getElementById('runtime-stop-btn').disabled = true;
      document.getElementById('runtime-status-dot').classList.remove('running');
      document.getElementById('runtime-status-text').textContent = 'Not running';
      document.getElementById('runtime-pid').textContent = '';
      
      // Stop polling
      stopRuntimePolling();
      
      // Hide WebUI
      const webui = document.getElementById('runtime-webui');
      const placeholder = document.getElementById('webui-placeholder');
      webui.classList.remove('active');
      webui.src = 'about:blank';
      placeholder.style.display = 'flex';
    } else {
      appendTerminalLine(`✗ Failed to stop: ${data.error}`, 'error');
    }
  } catch (error) {
    appendTerminalLine(`✗ Error: ${error.message}`, 'error');
  }
}

async function checkRuntimeStatus() {
  try {
    const result = await RuntimeGetStatus();
    const data = JSON.parse(result);
    
    if (data.running) {
      document.getElementById('runtime-start-btn').disabled = true;
      document.getElementById('runtime-stop-btn').disabled = false;
      document.getElementById('runtime-status-dot').classList.add('running');
      document.getElementById('runtime-status-text').textContent = 'Running';
      document.getElementById('runtime-pid').textContent = `PID: ${data.pid}`;
      
      if (!runtimePollingInterval) {
        startRuntimePolling();
      }
      
      // Load WebUI if not already loaded
      const webui = document.getElementById('runtime-webui');
      if (!webui.classList.contains('active')) {
        loadWebUI();
      }
    } else {
      document.getElementById('runtime-start-btn').disabled = false;
      document.getElementById('runtime-stop-btn').disabled = true;
      document.getElementById('runtime-status-dot').classList.remove('running');
      document.getElementById('runtime-status-text').textContent = 'Not running';
      document.getElementById('runtime-pid').textContent = '';
    }
  } catch (error) {
    console.error('Failed to check runtime status:', error);
  }
}

function startRuntimePolling() {
  if (runtimePollingInterval) {
    clearInterval(runtimePollingInterval);
  }
  
  runtimePollingInterval = setInterval(async () => {
    await pollRuntimeOutput();
    await checkRuntimeStatus();
  }, 1000);
}

function stopRuntimePolling() {
  if (runtimePollingInterval) {
    clearInterval(runtimePollingInterval);
    runtimePollingInterval = null;
  }
}

async function pollRuntimeOutput() {
  try {
    const output = await RuntimeGetOutput();
    if (output && output.trim()) {
      const terminal = document.getElementById('terminal-output');
      const lines = output.split('\n');
      
      // Only add new lines
      const currentLines = terminal.querySelectorAll('.terminal-line');
      const currentCount = currentLines.length;
      
      if (lines.length > currentCount) {
        const newLines = lines.slice(currentCount);
        newLines.forEach(line => {
          if (line.trim()) {
            appendTerminalLine(line);
          }
        });
      }
    }
  } catch (error) {
    console.error('Failed to poll runtime output:', error);
  }
}

function appendTerminalLine(text, type = '') {
  const terminal = document.getElementById('terminal-output');
  const line = document.createElement('div');
  line.className = `terminal-line ${type}`;
  
  // Add timestamp
  const time = new Date().toLocaleTimeString();
  line.innerHTML = `<span class="terminal-prompt">[${time}]</span> ${escapeHtml(text)}`;
  
  terminal.appendChild(line);
  terminal.scrollTop = terminal.scrollHeight;
}

function escapeHtml(text) {
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
}

function clearTerminal() {
  const terminal = document.getElementById('terminal-output');
  terminal.innerHTML = '<div class="terminal-welcome"><span class="terminal-prompt">$</span> Terminal cleared.</div>';
}

function loadWebUI() {
  const webui = document.getElementById('runtime-webui');
  const placeholder = document.getElementById('webui-placeholder');
  
  webui.src = 'http://localhost:6185';
  webui.classList.add('active');
  placeholder.style.display = 'none';
  
  appendTerminalLine('WebUI loaded at http://localhost:6185', 'success');
}
