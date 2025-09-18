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
const printerSelect = document.getElementById('printerSelect');

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
    refreshSetting();
});

// 更新设置
function refreshSetting() {
    const printer = printerSelect.value;
    const pageSize = pageSizeSelect.value;
    const orientation = orientationSelect.value;
    fetch('/setting',{
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            printer: printer,
            pageSize: pageSize,
            orientation: orientation
        })
    }).then((response) => {
        if (!response.ok) {
            throw new Error('网络响应不正常');
        }
        return response.json();
    }).then(data => {
        console.log('更新设置:', data);
        // showNotification(data.message || '打印任务已发送');
    })
}

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
    const printer = printerSelect.value;
    const pageSize = pageSizeSelect.value;
    const orientation = orientationSelect.value;

    if (!printer) {
        showNotification('请先选择打印机', 'error');
        return;
    }
    // 禁用按钮防止重复点击
    const printBtn = document.querySelector(`.print-btn[data-filename="${filename}"]`);
    if (printBtn) printBtn.disabled = true;

    fetch('/print', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            printer: printer,
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
    const printer = printerSelect.value;
    const pageSize = pageSizeSelect.value;
    const orientation = orientationSelect.value;

    fetch('/print_all', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            printer: printer,
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


// 打印机管理功能
document.addEventListener('DOMContentLoaded', function() {
    const printerSelect = document.getElementById('printerSelect');
    const refreshPrintersBtn = document.getElementById('refreshPrinters');

    // 加载打印机列表
    loadPrinters().then(r => {
        console.log("加载打印机列表完成!!!")
    });

    // 刷新打印机列表
    refreshPrintersBtn.addEventListener('click', loadPrinters);

    // 打印机选择变化时保存到本地存储
    printerSelect.addEventListener('change', function() {
        localStorage.setItem('selectedPrinter', this.value);
    });
});

// 加载打印机列表
async function loadPrinters() {
    const printerSelect = document.getElementById('printerSelect');
    const refreshPrintersBtn = document.getElementById('refreshPrinters');

    try {
        printerSelect.innerHTML = '<option value="">正在加载打印机...</option>';
        refreshPrintersBtn.disabled = true;

        const response = await fetch('/get_printers');
        if (!response.ok) {
            throw new Error('获取打印机列表失败');
        }

        const printers = await response.json();

        if (printers.length === 0) {
            printerSelect.innerHTML = '<option value="">未找到打印机</option>';
            return;
        }

        // 清空并填充打印机选项
        printerSelect.innerHTML = '';
        printers.forEach(printer => {
            const option = document.createElement('option');
            option.value = printer;
            option.textContent = printer;
            printerSelect.appendChild(option);
        });

        // 恢复之前选择的打印机
        const savedPrinter = localStorage.getItem('selectedPrinter');
        if (savedPrinter && printers.includes(savedPrinter)) {
            printerSelect.value = savedPrinter;
        } else if (printers.length > 0) {
            printerSelect.value = printers[0];
        }

    } catch (error) {
        console.error('加载打印机失败:', error);
        printerSelect.innerHTML = '<option value="">加载失败，点击刷新</option>';
    } finally {
        refreshPrintersBtn.disabled = false;
    }
}


// 刷新二维码
generateQRBtn.addEventListener('click', function () {
    document.getElementById('qrImage').src = '/qrcode?' + new Date().getTime();
    showNotification('二维码已刷新');
});
// 初始加载
document.addEventListener('DOMContentLoaded', function () {
    updateStatus('正在连接...', 'waiting');
    refreshQueue();
    refreshSetting();

    // 每30秒刷新一次
    setInterval(refreshQueue, 30000);
    setInterval(refreshSetting, 30000);
});