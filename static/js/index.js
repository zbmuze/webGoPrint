// 获取DOM元素
const statusIndicator = document.getElementById('statusIndicator');
const statusText = document.getElementById('statusText');
const serverAddress = document.getElementById('serverAddress');
const queueContainer = document.getElementById('queue');
const generateQRBtn = document.getElementById('generateQR');
const resetBtn = document.getElementById('reset');
const printAllBtn = document.getElementById('printAll');
const clearQueueBtn = document.getElementById('clearQueue');
const notification = document.getElementById('notification');
const pageSizeSelect = document.getElementById('pageSize');
const orientationSelect = document.getElementById('orientation');
const refreshQueueBtn = document.getElementById('refreshQueueBtn');
// 设置服务器地址
serverAddress.textContent = window.location.host;

// 显示通知
function showNotification(message, type = 'success') {
    notification.textContent = message;
    notification.className = `notification ${type}`;
    notification.classList.add('show');

    setTimeout(() => {
        notification.classList.remove('show');
    }, 3000);
}

// 更新状态指示器
function updateStatus(status, type) {
    statusText.textContent = status;
    statusIndicator.className = 'status-indicator';

    switch (type) {
        case 'online':
            statusIndicator.classList.add('status-online');
            break;
        case 'offline':
            statusIndicator.classList.add('status-offline');
            break;
        case 'waiting':
            statusIndicator.classList.add('status-waiting');
            break;
    }
}

// 获取文件图标类
function getFileIconClass(filename) {
    const ext = filename.split('.').pop().toLowerCase();
    return `file-icon ${ext}`;
}
// 主动触发队列刷新）
refreshQueueBtn.addEventListener('click', function () {
    showNotification('正在刷新队列...', 'warning'); // 提示用户
    refreshQueue(); // 直接调用函数，主动触发刷新
});

// 刷新队列函数
function refreshQueue() {
    fetch('/queue')
        .then(response => {
            if (!response.ok) {
                throw new Error('网络响应不正常');
            }
            return response.json();
        })
        .then(data => {
            if (data.files && data.files.length === 0) {
                queueContainer.innerHTML = '<div class="status">打印队列为空</div>';
                updateStatus('等待文件上传', 'waiting');
                return;
            }

            let queueHTML = '';
            data.files.forEach(file => {
                const ext = file.name.split('.').pop().toUpperCase();
                // 安全处理文件名，防止XSS攻击
                const safeFileName = file.name.replace(/"/g, '&quot;').replace(/'/g, '&#x27;');
                const iconClass = getFileIconClass(file.name);

                queueHTML += `
                        <div class="queue-item">
                            <div class="file-info">
                                <div class="${iconClass}">${ext}</div>
                                <div>
                                    <div>${safeFileName}</div>
                                    <div class="file-details">${file.size} · ${file.upload_time}</div>
                                </div>
                            </div>
                            <button class="print-btn" data-filename="${safeFileName}">打印</button>
                        </div>
                    `;
            });

            queueContainer.innerHTML = queueHTML;

            // 添加打印按钮事件监听
            document.querySelectorAll('.print-btn').forEach(btn => {
                btn.addEventListener('click', function () {
                    const filename = this.getAttribute('data-filename');
                    printFile(filename);
                });
            });

            updateStatus(`${data.files.length}个文件等待打印`, 'online');
        })
        .catch(error => {
            console.error('获取队列错误:', error);
            queueContainer.innerHTML = '<div class="status">加载队列失败</div>';
            updateStatus('连接错误', 'offline');
        });
}

// 打印文件
function printFile(filename) {
    console.log('尝试打印文件:', filename);

    // 获取打印设置
    const pageSize = pageSizeSelect.value;
    const orientation = orientationSelect.value;

    // 禁用按钮防止重复点击
    const printBtn = document.querySelector(`.print-btn[data-filename="${filename}"]`);
    if (printBtn) printBtn.disabled = true;

    fetch('/print', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            filename: filename,
            pageSize: pageSize,
            orientation: orientation
        })
    })
        .then(response => {
            if (!response.ok) {
                return response.json().then(err => {
                    throw new Error(err.error || '打印失败: HTTP ' + response.status)
                });
            }
            return response.json();
        })
        .then(data => {
            console.log('打印响应:', data);
            showNotification(data.message || '打印任务已发送');
            refreshQueue();
        })
        .catch(error => {
            console.error('打印错误详情:', error);
            showNotification('打印失败: ' + error.message, 'error');
            // 重新启用按钮
            if (printBtn) printBtn.disabled = false;
        });
}

// 打印全部
printAllBtn.addEventListener('click', function () {
    const pageSize = pageSizeSelect.value;
    const orientation = orientationSelect.value;

    fetch('/print_all', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            pageSize: pageSize,
            orientation: orientation
        })
    })
        .then(response => {
            if (!response.ok) {
                return response.json().then(err => { throw new Error(err.error || '打印失败') });
            }
            return response.json();
        })
        .then(data => {
            showNotification(data.message || '所有打印任务已发送');
            refreshQueue();
        })
        .catch(error => {
            console.error('打印全部错误:', error);
            showNotification('打印失败: ' + error.message, 'error');
        });
});

// 清空队列
clearQueueBtn.addEventListener('click', function () {
    if (confirm('确定要清空打印队列吗？')) {
        fetch('/clear_queue', {
            method: 'POST'
        })
            .then(response => {
                if (!response.ok) {
                    return response.json().then(err => { throw new Error(err.error || '清空失败') });
                }
                return response.json();
            })
            .then(data => {
                showNotification(data.message || '打印队列已清空');
                refreshQueue();
            })
            .catch(error => {
                console.error('清空队列错误:', error);
                showNotification('清空失败: ' + error.message, 'error');
            });
    }
});

// 刷新二维码
generateQRBtn.addEventListener('click', function () {
    document.getElementById('qrImage').src = '/qrcode?' + new Date().getTime();
    showNotification('二维码已刷新');
});

// 重置系统
resetBtn.addEventListener('click', function () {
    if (confirm('确定要重置系统吗？所有上传的文件将被删除')) {
        fetch('/reset', {
            method: 'POST'
        })
            .then(response => {
                if (!response.ok) {
                    return response.json().then(err => { throw new Error(err.error || '重置失败') });
                }
                return response.json();
            })
            .then(data => {
                showNotification(data.message || '系统已重置');
                refreshQueue();
            })
            .catch(error => {
                console.error('重置错误:', error);
                showNotification('重置失败: ' + error.message, 'error');
            });
    }
});

// 初始加载
document.addEventListener('DOMContentLoaded', function () {
    updateStatus('正在连接...', 'waiting');
    refreshQueue();

    // 每30秒刷新一次队列
    setInterval(refreshQueue, 30000);
});