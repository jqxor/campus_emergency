<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'

const mobileSidebarOpen = ref(false)

const closeMobileSidebar = () => {
  mobileSidebarOpen.value = false
}

const toggleMobileSidebar = () => {
  mobileSidebarOpen.value = !mobileSidebarOpen.value
}

let onGlobalClick: ((ev: Event) => void) | null = null
let onResize: (() => void) | null = null

onMounted(async () => {
  try {
    const env: any = (import.meta as any).env || {}
    const setIf = (id: string, value: any) => {
      const el = document.getElementById(id) as HTMLInputElement | null
      if (el && value) el.value = String(value)
    }
    setIf('baseNav', env.VITE_NAV_BASE)
    setIf('basePlan', env.VITE_PLAN_BASE)
    setIf('baseRole', env.VITE_RBAC_BASE)

    await import('../app.js')

    onGlobalClick = (ev: Event) => {
      if (window.innerWidth > 980) return
      const target = ev.target as HTMLElement | null
      if (!target) return
      if (target.closest('.menu-btn[data-target], .quick[data-jump], .menu-btn--mini')) {
        closeMobileSidebar()
      }
    }
    document.addEventListener('click', onGlobalClick)

    onResize = () => {
      if (window.innerWidth > 980) closeMobileSidebar()
    }
    window.addEventListener('resize', onResize)

    if ('serviceWorker' in navigator) {
      navigator.serviceWorker.register('/sw.js').catch(() => {})
    }
  } catch (err: any) {
    const el = document.getElementById('output')
    if (el) el.textContent = '初始化失败: ' + (err?.message || String(err))
  }
})

onBeforeUnmount(() => {
  if (onGlobalClick) {
    document.removeEventListener('click', onGlobalClick)
    onGlobalClick = null
  }
  if (onResize) {
    window.removeEventListener('resize', onResize)
    onResize = null
  }
})
</script>

<template>
  <div class="noise"></div>
  <div class="layout" :class="{ 'is-mobile-open': mobileSidebarOpen }">
    <button class="mobile-nav-toggle" type="button" @click="toggleMobileSidebar" :aria-expanded="mobileSidebarOpen">
      {{ mobileSidebarOpen ? '收起菜单' : '展开菜单' }}
    </button>
    <div class="mobile-backdrop" @click="closeMobileSidebar" aria-hidden="true"></div>
    <aside class="sidebar" :class="{ 'sidebar-mobile-open': mobileSidebarOpen }">
      <div class="brand">
        <strong>校园智能路径优化与应急疏散系统</strong>
        <span>控制台 / 本地联调 / 可视化演示</span>
      </div>

      <div class="side-user">
        <div class="avatar" aria-hidden="true"></div>
        <div class="side-user-meta">
          <div class="side-user-name">示例用户</div>
          <div class="side-user-sub">teacher · 在线</div>
        </div>
        <button class="menu-btn menu-btn--mini" data-target="viewUser" type="button">用户中心</button>
      </div>

      <nav class="menu" aria-label="主导航">
        <button class="menu-btn active" data-target="viewHome" type="button">首页概览</button>
        <button class="menu-btn" data-target="viewPath" type="button">智能路径优化</button>
        <button class="menu-btn" data-target="viewEmergency" type="button">应急疏散预案</button>
        <button class="menu-btn" data-target="viewReport" type="button">统计报表</button>
        <button class="menu-btn" data-target="viewMonitor" type="button">监控分析</button>
        <button class="menu-btn" data-target="viewSystem" type="button">系统管理</button>
      </nav>

      <div class="side-foot">
        <div class="pill">提示：底部“输出”面板显示接口响应</div>
      </div>
    </aside>

    <main class="main">
      <div style="display:none" aria-hidden="true">
        <input id="baseNav" value="http://localhost:8080" />
        <input id="basePlan" value="http://localhost:8081" />
        <input id="baseRole" value="http://localhost:8082" />
      </div>

      <section id="viewLogin" class="view card view--login" style="display:none" aria-hidden="true">
        <div class="auth-shell">
          <div class="auth-hero" aria-hidden="true">
            <div class="auth-hero-brand">
              <div class="auth-hero-kicker">Campus Safety Console</div>
              <div class="auth-hero-title">欢迎回来</div>
              <div class="auth-hero-sub">登录后进入后台控制台；不同角色自动隔离功能入口。</div>
            </div>
            <div class="auth-hero-list">
              <div class="auth-hero-item"><span class="dot"></span><span>智能路径优化与应急联动</span></div>
              <div class="auth-hero-item"><span class="dot"></span><span>RBAC 权限隔离（前端 + 后端）</span></div>
              <div class="auth-hero-item"><span class="dot"></span><span>本地联调：输出面板统一展示响应</span></div>
            </div>
          </div>

          <div class="auth-panel">
            <div class="view-head">
              <div>
                <h2>统一认证入口</h2>
                <p>登录与注册已拆分为独立流程，便于后续对接真实认证服务。</p>
              </div>
              <div class="view-badge">AUTH</div>
            </div>

            <div class="subtabs auth-subtabs">
              <button class="subtab-btn active auth-tab-btn" data-auth-tab="authLogin" type="button">登录</button>
              <button class="subtab-btn auth-tab-btn" data-auth-tab="authRegister" type="button">注册</button>
            </div>

            <div id="authLogin" class="auth-tab-panel active">
              <div class="grid cols-2 auth-grid">
                <div class="card">
                  <h3>账号登录</h3>
                  <label>
                    账号或邮箱
                    <input id="loginUser" placeholder="输入账号或邮箱" />
                  </label>
                  <label>
                    密码
                    <input id="loginPass" type="password" placeholder="******" />
                  </label>
                  <div class="inline auth-remember">
                    <input id="rememberMe" type="checkbox" />
                    <span>记住我（本地）</span>
                  </div>
                  <div class="row-actions">
                    <button id="btnLogin" type="button">登录</button>
                  </div>
                </div>
                <div class="card">
                  <h3>验证码登录</h3>
                  <label>
                    手机号
                    <input id="loginPhone" placeholder="13800000000" />
                  </label>
                  <div class="row-actions">
                    <button id="btnGetCode" class="ghost" type="button">获取验证码</button>
                    <button id="btnCodeLogin" type="button">验证码登录</button>
                  </div>
                  <div class="callout">验证码演示值：<span class="mono">123456</span></div>
                </div>
              </div>
            </div>

            <div id="authRegister" class="auth-tab-panel">
              <div class="card">
                <h3>账号注册</h3>
                <label>
                  姓名
                  <input id="registerName" placeholder="请输入姓名" />
                </label>
                <label>
                  学号
                  <input id="registerStudentId" placeholder="请输入学号" />
                </label>
                <label>
                  用户名
                  <input id="registerUser" placeholder="3-16位字母/数字/下划线" />
                </label>
                <label>
                  邮箱
                  <input id="registerEmail" placeholder="student001@ncist.edu.cn" />
                </label>
                <label>
                  密码
                  <input id="registerPass" type="password" placeholder="至少8位，包含字母和数字" />
                </label>
                <label>
                  确认密码
                  <input id="registerConfirmPass" type="password" placeholder="再次输入密码" />
                </label>
                <label>
                  角色
                  <select id="registerRole">
                    <option value="student">学生</option>
                    <option value="teacher">教师</option>
                  </select>
                </label>
                <div class="guideline-box">
                  <h4>用户准则</h4>
                  <p>本系统用于校园安全教学、演示与应急联动训练，不得用于未授权测试或攻击。</p>
                  <ul>
                    <li>仅在合法授权环境下使用。</li>
                    <li>不得将演示流程用于真实破坏行为。</li>
                    <li>需遵守校内与国家网络安全相关法规。</li>
                  </ul>
                </div>
                <label class="guideline-check">
                  <input id="registerAgree" type="checkbox" />
                  <span>我已阅读并同意以上用户准则</span>
                </label>
                <div class="row-actions">
                  <button id="btnRegister" type="button">提交注册</button>
                </div>
              </div>
            </div>

            <div id="authError" class="auth-error" aria-live="polite"></div>
          </div>
        </div>
      </section>

      <section id="viewUser" class="view card view--user">
        <div class="view-head">
          <div>
            <h2>用户中心</h2>
            <p>基础用户界面：个人信息、常用入口、操作留痕（演示）。</p>
          </div>
          <div class="view-badge">USER</div>
        </div>
        <div class="grid cols-2">
          <div class="card">
            <h3>个人信息</h3>
            <div class="profile">
              <div class="avatar avatar--lg" aria-hidden="true"></div>
              <div>
                <div class="profile-name">示例用户</div>
                <div class="profile-sub">teacher · 校园安全联动组</div>
                <div class="chips">
                  <span class="chip">RBAC 已启用</span>
                  <span class="chip">本地联调</span>
                  <span class="chip">只读演示</span>
                </div>
              </div>
            </div>
            <p class="helper">真实项目中可在此对接用户资料、令牌刷新、二次认证等。</p>
          </div>
          <div class="card">
            <h3>常用入口</h3>
            <div class="quick-grid quick-grid--user">
              <a class="card quick" data-jump="viewPath" href="#viewPath">
                <span class="quick-title">路径计算</span>
                <small class="quick-meta">规划最优通行路线</small>
              </a>
              <a class="card quick" data-jump="viewEmergency" href="#viewEmergency">
                <span class="quick-title">预案检索</span>
                <small class="quick-meta">按场景快速定位预案</small>
              </a>
              <a class="card quick" data-jump="viewMonitor" href="#viewMonitor">
                <span class="quick-title">预警测试</span>
                <small class="quick-meta">触发实时预警流程</small>
              </a>
              <a class="card quick" data-jump="viewSystem" href="#viewSystem">
                <span class="quick-title">用户管理</span>
                <small class="quick-meta">维护账户与权限</small>
              </a>
              <a class="card quick" data-jump="viewReport" href="#viewReport">
                <span class="quick-title">报表导出</span>
                <small class="quick-meta">生成日报与周报</small>
              </a>
            </div>
          </div>
        </div>
      </section>

      <section id="viewHome" class="view card active view--home">
        <div class="view-head">
          <div>
            <h2>首页概览</h2>
            <p>核心指标与快捷入口（指标将尝试从后端拉取并自动刷新）。</p>
          </div>
          <div class="view-badge">DASH</div>
        </div>

        <div class="stats-grid">
          <div class="card stat">
            <span>已完成导航次数</span>
            <strong id="metricPath">-</strong>
          </div>
          <div class="card stat">
            <span>预案总数</span>
            <strong id="metricPlan">-</strong>
          </div>
          <div class="card stat">
            <span>角色总数</span>
            <strong id="metricRole">-</strong>
          </div>
          <div class="card stat">
            <span>系统状态</span>
            <strong>OK</strong>
          </div>
        </div>

        <div class="home-grid">
          <div class="card">
            <h3>今日重点</h3>
            <div class="timeline">
              <div class="tl-item"><span class="dot"></span><span>检查教学区高峰期拥堵点</span></div>
              <div class="tl-item"><span class="dot"></span><span>核对应急预案状态与负责人</span></div>
              <div class="tl-item"><span class="dot"></span><span>抽测监控预警与日志导出</span></div>
            </div>
            <p class="helper">这是演示内容，真实项目可接入值班计划与事件中心。</p>
          </div>
          <div class="card">
            <h3>快捷操作</h3>
            <div class="quick-grid">
              <a class="card quick" data-jump="viewPath" href="#viewPath">
                <span class="quick-title">路径计算</span>
                <small class="quick-meta">避障 + 导航联动</small>
              </a>
              <a class="card quick" data-jump="viewEmergency" href="#viewEmergency">
                <span class="quick-title">预案检索</span>
                <small class="quick-meta">查看状态与责任人</small>
              </a>
              <a class="card quick" data-jump="viewSystem" href="#viewSystem">
                <span class="quick-title">用户管理</span>
                <small class="quick-meta">角色权限快速配置</small>
              </a>
              <a class="card quick" data-jump="viewMonitor" href="#viewMonitor">
                <span class="quick-title">预警测试</span>
                <small class="quick-meta">模拟告警与通知</small>
              </a>
              <a class="card quick" data-jump="viewReport" href="#viewReport">
                <span class="quick-title">报表导出</span>
                <small class="quick-meta">一键输出汇总报告</small>
              </a>
            </div>
          </div>
        </div>
      </section>

      <section id="viewPath" class="view card view--path">
        <div class="view-head">
          <div>
            <h2>智能路径优化</h2>
            <p>导航路径计算、障碍物管理、历史与摘要。</p>
          </div>
          <div class="view-badge">NAV</div>
        </div>

        <div class="subtabs">
          <button class="subtab-btn active" data-subtab="pathNav" type="button">导航接口</button>
          <button class="subtab-btn" data-subtab="pathObs" type="button">障碍物</button>
          <button class="subtab-btn" data-subtab="pathAna" type="button">分析导出</button>
        </div>

        <div id="pathNav" class="subtab active">
          <div id="pathMapShell" class="map-center-shell">
            <div class="card map-stage-card">
              <h3>华北科技学院卫星路径标注</h3>
              <div id="pathMap" class="map-canvas map-canvas--focus"></div>
              <div class="row-actions">
                <button id="btnPathRenderMap" type="button">在卫星图标注路径</button>
                <button id="btnPathE2ENav" type="button">端到端导航</button>
                <button id="btnPathClearMap" class="ghost" type="button">清空标注</button>
                <button id="btnPathRequestGps" class="ghost" type="button">申请GPS权限</button>
                <button id="btnPathLocate" class="ghost" type="button">定位到我</button>
                <button id="btnPathResetToMe" class="ghost" type="button">重置到我位置</button>
              </div>
              <div class="row-actions">
                <button id="btnPathCenterPin" class="ghost" type="button">中心打点</button>
                <button id="btnPathMeasureToMe" class="ghost" type="button">测距到我</button>
                <button id="btnPathFitCampus" class="ghost" type="button">回到校园</button>
              </div>
            </div>

            <aside class="card map-side-panel">
              <div class="map-side-head">
                <h3>路径控制侧栏</h3>
                <button id="btnPathSidebarToggle" class="ghost" type="button">收起侧栏</button>
              </div>
              <div class="map-side-content">
                <div class="card card-flat">
                  <h3>算法参数</h3>
                  <label>
                    算法
                    <select id="algoType">
                      <option value="dijkstra">Dijkstra</option>
                      <option value="astar">A*</option>
                    </select>
                  </label>
                  <label>
                    权重：距离
                    <input id="weightDistance" type="number" value="1" />
                  </label>
                  <label>
                    权重：时间
                    <input id="weightTime" type="number" value="1" />
                  </label>
                  <label>
                    权重：安全
                    <input id="weightSafe" type="number" value="1" />
                  </label>
                  <div class="row-actions">
                    <button id="btnSaveAlgo" type="button">保存参数</button>
                    <button id="btnPathSimulation" class="ghost" type="button">测试模拟</button>
                  </div>
                </div>

                <div class="card card-flat">
                  <h3>导航请求</h3>
                  <label>
                    X-User-ID
                    <input id="navUserId" value="1001" />
                  </label>
                  <div class="grid cols-2">
                    <label>
                      开始日期（YYYY-MM-DD）
                      <input id="navStartDate" placeholder="2026-04-01" />
                    </label>
                    <label>
                      结束日期（YYYY-MM-DD）
                      <input id="navEndDate" placeholder="2026-04-19" />
                    </label>
                  </div>
                  <label>
                    Path ID
                    <input id="navPathId" value="1" />
                  </label>
                  <label>
                    请求体（JSON）
                    <textarea id="navCalcPayload">{
  "start_lat": 39.9515268,
  "start_lng": 116.7986691,
  "end_lat": 39.9572022,
  "end_lng": 116.7989527,
  "algorithm": "astar"
}</textarea>
                  </label>
                  <div class="row-actions">
                    <button id="btnNavCalc" type="button">计算路径</button>
                    <button id="btnNavStart" class="ghost" type="button">开始导航</button>
                    <button id="btnNavUpdate" class="ghost" type="button">更新</button>
                    <button id="btnNavEnd" class="ghost" type="button">结束</button>
                  </div>
                  <label>
                    端到端导航模式
                    <select id="pathNavMode">
                      <option value="car">驾车</option>
                      <option value="walk">步行</option>
                    </select>
                  </label>
                  <label>
                    终点（坐标：纬度,经度）
                    <input id="pathDestinationInput" placeholder="例如 39.9572,116.7990" />
                  </label>
                  <div class="row-actions">
                    <button id="btnPathPickDestination" class="ghost" type="button">地图点击选终点</button>
                    <button id="btnPathLiveNavStop" class="danger" type="button">停止实时导航</button>
                  </div>
                  <div id="pathLiveNavInfo" class="log">实时导航未开始</div>
                  <div class="row-actions">
                    <button id="btnNavHistoryExport" class="ghost" type="button">导出历史</button>
                    <button id="btnNavSummary" class="ghost" type="button">摘要</button>
                  </div>
                  <div class="row-actions">
                    <label class="grow">
                      Warning ID
                      <input id="warningId" value="1" />
                    </label>
                  </div>
                  <div class="row-actions">
                    <button id="btnWarningConfirm" class="ghost" type="button">确认障碍</button>
                    <button id="btnWarningIgnore" class="ghost" type="button">忽略障碍</button>
                  </div>
                </div>
              </div>
            </aside>
          </div>
        </div>

        <div id="pathObs" class="subtab">
          <div class="grid cols-2">
            <div class="card">
              <h3>新增障碍物</h3>
              <label>
                类型
                <select id="obType">
                  <option value="roadblock">路障</option>
                  <option value="construction">施工</option>
                  <option value="crowd">人群拥堵</option>
                </select>
              </label>
              <label>
                位置
                <input id="obLocation" placeholder="教学楼南门" />
              </label>
              <p class="helper">若要参与地图避障，请用坐标格式输入：纬度,经度（如 39.9562,116.7972）。</p>
              <label>
                影响范围
                <input id="obRange" placeholder="50m" />
              </label>
              <div class="row-actions">
                <button id="btnAddObstacle" type="button">新增</button>
              </div>
              <div class="row-actions">
                <input id="obCsvFile" type="file" accept=".csv" class="grow" />
                <button id="btnImportObstacleCsv" class="ghost" type="button">导入CSV</button>
              </div>
            </div>
            <div class="card">
              <h3>障碍物列表</h3>
              <table>
                <thead>
                  <tr>
                    <th>位置</th>
                    <th>类型</th>
                    <th>状态</th>
                  </tr>
                </thead>
                <tbody id="obstacleTable"></tbody>
              </table>
            </div>
          </div>
        </div>

        <div id="pathAna" class="subtab">
          <div class="grid cols-2">
            <div class="card">
              <h3>导入 / 导出</h3>
              <div class="row-actions">
                <input id="pathCsvFile" type="file" accept=".csv" class="grow" />
                <button id="btnImportPathCsv" class="ghost" type="button">导入CSV</button>
                <button id="btnExportPathCsv" class="ghost" type="button">导出CSV</button>
              </div>
              <p class="helper">导入/导出为演示流程；后续可对接后端配置接口。</p>
            </div>
            <div class="card">
              <h3>分析报告</h3>
              <div class="row-actions">
                <button id="btnPathAnalyze" type="button">效率分析</button>
                <button id="btnPathAdvice" class="ghost" type="button">优化建议</button>
                <button id="btnPathReportExport" class="ghost" type="button">导出报告</button>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section id="viewEmergency" class="view card view--emergency">
        <div class="view-head">
          <div>
            <h2>应急疏散预案</h2>
            <p>预案 CRUD、检索、模拟、事件触发、优化。</p>
          </div>
          <div class="view-badge">PLAN</div>
        </div>

        <div class="subtabs">
          <button class="subtab-btn active" data-subtab="planCrud" type="button">预案管理</button>
          <button class="subtab-btn" data-subtab="planSim" type="button">实时模拟</button>
          <button class="subtab-btn" data-subtab="planEvent" type="button">事件触发</button>
          <button class="subtab-btn" data-subtab="planOpt" type="button">路径优化</button>
        </div>

        <div id="planCrud" class="subtab active">
          <div class="grid cols-2">
            <div class="card">
              <h3>预案请求体（JSON）</h3>
              <label>
                预案 ID
                <input id="planId" value="1" />
              </label>
              <label>
                Body
                <textarea id="planCreatePayload">{
  "name": "食堂火灾疏散",
  "scenario_type": "fire",
  "status": "draft",
  "description": "演示预案"
}</textarea>
              </label>
              <div class="row-actions">
                <button id="btnPlanCreate" type="button">创建</button>
                <button id="btnPlanUpdate" class="ghost" type="button">更新</button>
                <button id="btnPlanDelete" class="danger" type="button">删除</button>
                <button id="btnPlanGet" class="ghost" type="button">获取</button>
              </div>
              <label>
                更新状态 Body
                <textarea id="planStatusBody">{
  "status": "active"
}</textarea>
              </label>
              <div class="row-actions">
                <button id="btnPlanUpdateStatus" class="ghost" type="button">更新状态</button>
              </div>
            </div>
            <div class="card">
              <h3>检索 / 导入 / 导出</h3>
              <label>
                场景类型
                <input id="planScenario" placeholder="fire" />
              </label>
              <label>
                状态
                <input id="planStatus" placeholder="active" />
              </label>
              <label>
                关键词
                <input id="planKeyword" placeholder="食堂" />
              </label>
              <div class="row-actions">
                <button id="btnPlanSearch" type="button">搜索</button>
                <button id="btnPlanExport" class="ghost" type="button">导出</button>
              </div>
              <div class="row-actions">
                <input id="planImportFile" type="file" class="grow" />
                <button id="btnPlanImport" class="ghost" type="button">导入</button>
              </div>
            </div>
          </div>
        </div>

        <div id="planSim" class="subtab">
          <div id="evacMapShell" class="map-center-shell">
            <div class="card map-stage-card">
              <h3>卫星疏散路线标注</h3>
              <div id="evacMap" class="map-canvas map-canvas--focus"></div>
              <div class="row-actions">
                <button id="btnEvacRenderMap" type="button">标注疏散路线</button>
                <button id="btnEvacE2ENav" type="button">端到端疏散导航</button>
                <button id="btnEvacClearMap" class="ghost" type="button">清空标注</button>
                <button id="btnEvacRequestGps" class="ghost" type="button">申请GPS权限</button>
                <button id="btnEvacLocate" class="ghost" type="button">定位到我</button>
                <button id="btnEvacResetToMe" class="ghost" type="button">重置到我位置</button>
              </div>
              <div class="row-actions">
                <button id="btnEvacCenterPin" class="ghost" type="button">中心打点</button>
                <button id="btnEvacMeasureToMe" class="ghost" type="button">测距到我</button>
                <button id="btnEvacFitCampus" class="ghost" type="button">回到校园</button>
              </div>
            </div>

            <aside class="card map-side-panel">
              <div class="map-side-head">
                <h3>疏散控制侧栏</h3>
                <button id="btnEvacSidebarToggle" class="ghost" type="button">收起侧栏</button>
              </div>
              <div class="map-side-content">
                <div class="card card-flat">
                  <h3>模拟参数</h3>
                  <label>
                    预案 ID
                    <input id="simPlanId" value="1" />
                  </label>
                  <label>
                    人数
                    <input id="simPeople" type="number" value="300" />
                  </label>
                  <label>
                    速度（m/s）
                    <input id="simSpeed" type="number" value="1.2" />
                  </label>
                  <div class="row-actions">
                    <button id="btnRunSimulation" type="button">开始模拟</button>
                    <button id="btnSimReport" class="ghost" type="button">导出报告</button>
                  </div>
                  <label>
                    端到端疏散导航模式
                    <select id="evacNavMode">
                      <option value="walk">步行</option>
                      <option value="car">驾车</option>
                    </select>
                  </label>
                  <label>
                    终点（坐标：纬度,经度）
                    <input id="evacDestinationInput" placeholder="例如 39.9542,116.8006" />
                  </label>
                  <div class="row-actions">
                    <button id="btnEvacPickDestination" class="ghost" type="button">地图点击选终点</button>
                    <button id="btnEvacLiveNavStop" class="danger" type="button">停止实时导航</button>
                  </div>
                  <div id="evacLiveNavInfo" class="log">实时导航未开始</div>
                  <progress id="simProgress" value="0" max="100"></progress>
                </div>
              </div>
            </aside>
          </div>
        </div>

        <div id="planEvent" class="subtab">
          <div class="grid cols-2">
            <div class="card">
              <h3>事件触发</h3>
              <label>
                事件类型
                <select id="eventType">
                  <option value="fire">火灾</option>
                  <option value="earthquake">地震</option>
                  <option value="stampede">踩踏</option>
                </select>
              </label>
              <label>
                预案 ID
                <input id="eventPlanId" value="1" />
              </label>
              <div class="row-actions">
                <button id="btnTriggerEvent" type="button">触发</button>
                <button id="btnSendNotice" class="ghost" type="button">发送通知</button>
              </div>
              <div id="eventLog" class="log"></div>
            </div>
            <div class="card">
              <h3>反馈</h3>
              <label>
                内容
                <textarea id="eventFeedback" placeholder="填写现场反馈..."></textarea>
              </label>
              <div class="row-actions">
                <button id="btnSubmitFeedback" type="button">提交反馈</button>
              </div>
            </div>
          </div>
        </div>

        <div id="planOpt" class="subtab">
          <div class="grid cols-2">
            <div class="card">
              <h3>优化计算</h3>
              <label>
                预案 ID
                <input id="optPlanId" value="1" />
              </label>
              <div class="row-actions">
                <button id="btnPlanOptimize" type="button">调用优化</button>
              </div>
              <p class="helper">调用后端：`POST /api/plans/:id/optimize`。</p>
            </div>
            <div class="card">
              <h3>应用权重</h3>
              <label>
                距离权重
                <input id="optDist" type="number" value="1" />
              </label>
              <label>
                安全权重
                <input id="optSafe" type="number" value="1" />
              </label>
              <div class="row-actions">
                <button id="btnApplyPath" type="button">应用</button>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section id="viewReport" class="view card view--report">
        <div class="view-head">
          <div>
            <h2>统计报表</h2>
            <p>路径使用统计与疏散效率报表（当前为前端演示导出）。</p>
          </div>
          <div class="view-badge">RPT</div>
        </div>
        <div class="subtabs">
          <button class="subtab-btn active" data-subtab="rptPath" type="button">路径报表</button>
          <button class="subtab-btn" data-subtab="rptEvac" type="button">疏散报表</button>
        </div>
        <div id="rptPath" class="subtab active">
          <div class="grid cols-2">
            <div class="card">
              <h3>路径使用统计</h3>
              <label>
                起始日期
                <input id="rpStart" placeholder="2026-04-01" />
              </label>
              <label>
                结束日期
                <input id="rpEnd" placeholder="2026-04-19" />
              </label>
              <label>
                报表类型
                <select id="rpType">
                  <option value="daily">日</option>
                  <option value="weekly">周</option>
                  <option value="monthly">月</option>
                </select>
              </label>
              <div class="row-actions">
                <button id="btnRptGenerate" type="button">生成</button>
                <button id="btnRptEvaluate" class="ghost" type="button">效率评估</button>
                <button id="btnRptExport" class="ghost" type="button">导出</button>
              </div>
            </div>
            <div class="card hint-card">
              <h3>说明</h3>
              <p class="helper">如需与导航服务摘要联动，可在“智能路径优化”模块调用 summary。</p>
            </div>
          </div>
        </div>
        <div id="rptEvac" class="subtab">
          <div class="grid cols-2">
            <div class="card">
              <h3>应急响应</h3>
              <label>
                事件类型
                <input id="evType" placeholder="fire" />
              </label>
              <div class="row-actions">
                <button id="btnEvReport" type="button">响应时间报告</button>
                <button id="btnEvExport" class="ghost" type="button">导出</button>
              </div>
            </div>
            <div class="card">
              <h3>疏散效率</h3>
              <label>
                场景
                <input id="evScene" placeholder="食堂" />
              </label>
              <div class="row-actions">
                <button id="btnEvSuccess" type="button">成功率评估</button>
                <button id="btnEvStartSim" class="ghost" type="button">执行模拟</button>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section id="viewMonitor" class="view card view--monitor">
        <div class="view-head monitor-head">
          <div>
            <div class="monitor-eyebrow">态势感知中心</div>
            <h2>监控分析</h2>
            <p>摄像头选择、追踪预警、风险预测与综合报告统一收束在同一屏，适合巡检和值班快速判断。</p>
          </div>
          <div class="view-badge">MON</div>
        </div>

        <div class="monitor-overview">
          <div class="card monitor-hero-card">
            <div class="monitor-hero-top">
              <span class="monitor-live-pill"><span></span>LIVE 监测</span>
              <span class="monitor-hero-meta">值守模式 · 自动巡航</span>
            </div>
            <h3>将摄像、预警、报表放进一张“作战桌面”</h3>
            <p>以监控为核心入口，把实时画面、风险阈值、日志追踪与分析报告串成连续工作流，减少值班员在多个页面之间来回切换。</p>
            <div class="monitor-tag-row">
              <span class="monitor-tag">视频联动</span>
              <span class="monitor-tag">异常追踪</span>
              <span class="monitor-tag">风险阈值</span>
              <span class="monitor-tag">报告归档</span>
            </div>
          </div>

          <div class="monitor-stat-grid">
            <article class="monitor-stat-card">
              <span>接入通道</span>
              <strong>12 路</strong>
              <small>楼宇、路口、广场统一管理</small>
            </article>
            <article class="monitor-stat-card is-hot">
              <span>当前预警</span>
              <strong>2 条</strong>
              <small>高亮显示最近 10 分钟内事件</small>
            </article>
            <article class="monitor-stat-card is-calm">
              <span>分析周期</span>
              <strong>5 分钟</strong>
              <small>适合巡检与自动复核节奏</small>
            </article>
          </div>
        </div>

        <div class="subtabs monitor-tabs">
          <button class="subtab-btn active" data-subtab="monVideo" type="button">视频监控</button>
          <button class="subtab-btn" data-subtab="monRisk" type="button">风险预警</button>
          <button class="subtab-btn" data-subtab="monReport" type="button">综合报告</button>
        </div>
        <div id="monVideo" class="subtab active">
          <div class="monitor-layout">
            <div class="card monitor-stage-card">
              <div class="monitor-card-head">
                <div>
                  <h3>画面控制</h3>
                  <p>把摄像头、区域与播放模式放在同一组控件里，减少值班切换成本。</p>
                </div>
                <span class="monitor-chip">实时接管</span>
              </div>
              <label>
                区域
                <input id="camArea" placeholder="教学楼" />
              </label>
              <label>
                摄像头 ID
                <input id="camId" value="1" />
              </label>
              <label>
                模式
                <select id="camMode">
                  <option value="realtime">实时</option>
                  <option value="record">回放</option>
                </select>
              </label>
              <div class="row-actions">
                <button id="btnCamSelect" type="button">切换</button>
                <button id="btnSnapshot" class="ghost" type="button">导出截图</button>
              </div>
            </div>
            <div class="monitor-stack">
              <div class="card monitor-stage-card">
                <div class="monitor-card-head">
                  <div>
                    <h3>路径追踪</h3>
                    <p>追踪人流或车辆，形成即时告警的第一条线索。</p>
                  </div>
                  <span class="monitor-chip monitor-chip--soft">追踪中</span>
                </div>
              <label>
                追踪类型
                <select id="trackType">
                  <option value="people">人流</option>
                  <option value="vehicle">车辆</option>
                </select>
              </label>
              <label>
                更新频率（秒）
                <input id="trackFreq" type="number" value="3" />
              </label>
              <div class="row-actions">
                <button id="btnTrackStart" type="button">启动追踪</button>
              </div>
              <div id="liveAlert" class="alert"></div>
              </div>
            </div>
          </div>
        </div>
        <div id="monRisk" class="subtab">
          <div class="monitor-layout monitor-layout--risk">
            <div class="card monitor-stage-card">
              <div class="monitor-card-head">
                <div>
                  <h3>风险预测设置</h3>
                  <p>预设阈值、周期和风险类型，先把规则收敛到一个清晰的控制面板。</p>
                </div>
                <span class="monitor-chip monitor-chip--warn">阈值策略</span>
              </div>
              <label>
                类型
                <select id="riskType">
                  <option value="crowd">拥堵</option>
                  <option value="fire">火灾</option>
                  <option value="weather">极端天气</option>
                </select>
              </label>
              <label>
                阈值
                <input id="riskThreshold" type="number" value="80" />
              </label>
              <label>
                周期（分钟）
                <input id="riskCycle" type="number" value="5" />
              </label>
              <div class="row-actions">
                <button id="btnRiskPredict" type="button">保存设置</button>
                <button id="btnRiskTest" class="ghost" type="button">预警测试</button>
                <button id="btnWarnExport" class="ghost" type="button">导出日志</button>
              </div>
            </div>
            <div class="card monitor-log-card">
              <div class="monitor-card-head">
                <div>
                  <h3>预警日志</h3>
                  <p>按时间顺序展示风险信号、地点与处理状态，方便复盘。</p>
                </div>
                <span class="monitor-chip monitor-chip--soft">Recent</span>
              </div>
              <table>
                <thead>
                  <tr>
                    <th>时间</th>
                    <th>地点</th>
                    <th>等级</th>
                    <th>状态</th>
                  </tr>
                </thead>
                <tbody id="warnLog"></tbody>
              </table>
            </div>
          </div>
        </div>
        <div id="monReport" class="subtab">
          <div class="monitor-layout monitor-layout--report">
            <div class="card monitor-stage-card">
              <div class="monitor-card-head">
                <div>
                  <h3>报告生成</h3>
                  <p>把日常、事件和范围参数收在一组面板里，输出更像正式报表。</p>
                </div>
                <span class="monitor-chip monitor-chip--accent">Export</span>
              </div>
              <label>
                类型
                <select id="mrType">
                  <option value="daily">日常</option>
                  <option value="incident">事件</option>
                </select>
              </label>
              <label>
                范围
                <input id="mrRange" placeholder="全校" />
              </label>
              <div class="row-actions">
                <button id="btnMrGenerate" type="button">生成</button>
                <button id="btnMrEdit" class="ghost" type="button">编辑模板</button>
                <button id="btnMrExport" class="ghost" type="button">导出</button>
              </div>
            </div>
            <div class="card hint-card monitor-note-card">
              <h3>说明</h3>
              <p class="helper">综合分析报告当前为前端示例输出，后续可接入模型/规则引擎。</p>
            </div>
          </div>
        </div>
      </section>

      <section id="viewSystem" class="view card view--system">
        <div class="view-head">
          <div>
            <h2>系统管理</h2>
            <p>用户管理、角色管理、权限导入与审计。</p>
          </div>
          <div class="view-badge">SYS</div>
        </div>

        <div class="subtabs">
          <button class="subtab-btn active" data-subtab="sysUsers" type="button">用户</button>
          <button class="subtab-btn" data-subtab="sysRoles" type="button">角色</button>
          <button class="subtab-btn" data-subtab="sysPerms" type="button">权限</button>
          <button class="subtab-btn" data-subtab="sysCampus" type="button">校园范围配置</button>
        </div>

        <div id="sysUsers" class="subtab active">
          <div class="grid cols-2">
            <div class="card">
              <h3>用户编辑</h3>
              <label>
                用户名
                <input id="uName" placeholder="teacher01" />
              </label>
              <label>
                密码
                <input id="uPass" placeholder="123456" />
              </label>
              <label>
                邮箱
                <input id="uEmail" placeholder="teacher01@campus.edu" />
              </label>
              <label>
                角色 ID
                <input id="uRole" type="number" value="1" />
              </label>
              <label>
                用户权限（逗号分隔）
                <input id="userPerm" placeholder="path.read,plan.write" />
              </label>
              <div class="row-actions">
                <button id="btnUserAdd" type="button">新增</button>
                <button id="btnUserEdit" class="ghost" type="button">更新</button>
                <button id="btnUserDelete" class="danger" type="button">删除</button>
                <button id="btnUserPerm" class="ghost" type="button">分配权限</button>
              </div>
            </div>

            <div class="card">
              <h3>用户列表</h3>
              <table>
                <thead>
                  <tr>
                    <th>用户名</th>
                    <th>邮箱</th>
                    <th>角色</th>
                  </tr>
                </thead>
                <tbody id="userTable"></tbody>
              </table>
            </div>
          </div>
        </div>

        <div id="sysRoles" class="subtab">
          <div class="grid cols-2">
            <div class="card">
              <h3>创建角色</h3>
              <label>
                Body
                <textarea id="roleCreatePayload">{
  "name": "teacher",
  "description": "教师角色"
}</textarea>
              </label>
              <div class="row-actions">
                <button id="btnRoleCreate" type="button">创建</button>
                <button id="btnRoleList" class="ghost" type="button">列表</button>
              </div>
              <label>
                Role ID
                <input id="roleId" type="number" value="1" />
              </label>
              <div class="row-actions">
                <button id="btnRoleGet" class="ghost" type="button">获取</button>
                <button id="btnRoleDelete" class="danger" type="button">删除</button>
                <button id="btnRoleExport" class="ghost" type="button">导出CSV</button>
              </div>
            </div>
            <div class="card">
              <h3>更新与授权</h3>
              <label>
                Update Body
                <textarea id="roleUpdateBody">{
  "name": "teacher",
  "description": "教师（更新）",
  "is_active": true
}</textarea>
              </label>
              <div class="row-actions">
                <button id="btnRoleUpdate" type="button">更新</button>
              </div>
              <label>
                Permission IDs（JSON 数组）
                <textarea id="rolePermIds">[1,2,3]</textarea>
              </label>
              <div class="row-actions">
                <button id="btnRoleAssignPerm" class="ghost" type="button">分配权限</button>
                <button id="btnRoleGetPerm" class="ghost" type="button">查看权限</button>
              </div>
            </div>
          </div>
        </div>

        <div id="sysPerms" class="subtab">
          <div class="grid cols-2">
            <div class="card">
              <h3>权限树 / 审计</h3>
              <div class="row-actions">
                <button id="btnPermAssign" type="button">查看权限树</button>
              </div>
              <div class="row-actions">
                <input id="permImport" type="file" class="grow" />
                <button id="btnPermImport" class="ghost" type="button">导入权限配置</button>
              </div>
              <p class="helper">导入后会自动请求审计日志。</p>
            </div>
            <div class="card hint-card">
              <h3>说明</h3>
              <p class="helper">角色/用户权限分配分别在“角色”“用户”页完成。</p>
            </div>
          </div>
        </div>

        <div id="sysCampus" class="subtab">
          <div id="adminCampusMapShell" class="map-center-shell">
            <div class="card map-stage-card">
              <h3>Admin 校园范围编辑（框选 + 点位维护）</h3>
              <div id="adminCampusMap" class="map-canvas map-canvas--focus"></div>
              <div class="row-actions">
                <button id="btnAdminStartBoundary" type="button">开始框选校园范围</button>
                <button id="btnAdminFinishBoundary" class="ghost" type="button">完成框选</button>
                <button id="btnAdminUndoBoundary" class="ghost" type="button">撤销最后一点</button>
                <button id="btnAdminClearBoundary" class="danger" type="button">清空范围</button>
              </div>
              <div class="row-actions">
                <label class="grow">
                  点位类型
                  <select id="adminPointType">
                    <option value="assembly">集合点</option>
                    <option value="evacuation">疏散点</option>
                    <option value="risk">风险点</option>
                    <option value="gate">校门</option>
                    <option value="building">建筑</option>
                  </select>
                </label>
                <label class="grow">
                  点位名称
                  <input id="adminPointName" placeholder="例如：北区集合点A" />
                </label>
              </div>
              <div class="row-actions">
                <button id="btnAdminStartPoint" type="button">点击地图新增点位</button>
                <button id="btnAdminStopEdit" class="ghost" type="button">停止编辑</button>
                <button id="btnAdminClearPoints" class="danger" type="button">清空全部点位</button>
              </div>
            </div>

            <aside class="card map-side-panel">
              <div class="map-side-head">
                <h3>校园配置工具</h3>
              </div>
              <div class="map-side-content">
                <div class="card card-flat">
                  <h3>保存 / 发布</h3>
                  <div class="row-actions">
                    <button id="btnAdminSaveCampusGeo" type="button">保存到本地配置</button>
                    <button id="btnAdminApplyCampusGeo" class="ghost" type="button">应用到路径/疏散地图</button>
                  </div>
                  <div class="row-actions">
                    <button id="btnAdminResetCampusGeo" class="danger" type="button">恢复默认校园配置</button>
                  </div>
                </div>

                <div class="card card-flat">
                  <h3>导入 / 导出</h3>
                  <label>
                    校园配置 JSON
                    <textarea id="adminCampusJson" placeholder='{"outline":[[39.95,116.79]],"points":[]}'></textarea>
                  </label>
                  <div class="row-actions">
                    <button id="btnAdminExportCampusGeo" class="ghost" type="button">导出JSON到文本框</button>
                    <button id="btnAdminImportCampusGeo" class="ghost" type="button">从文本框导入</button>
                  </div>
                </div>

                <div class="card card-flat">
                  <h3>辅助功能</h3>
                  <div class="row-actions">
                    <button id="btnAdminFindNearestAssembly" class="ghost" type="button">计算最近集合点</button>
                    <button id="btnAdminGenerateDraftRoute" class="ghost" type="button">生成演练疏散草案</button>
                  </div>
                  <div id="adminGeoSummary" class="log">尚未加载校园配置</div>
                </div>
              </div>
            </aside>
          </div>
        </div>
      </section>

      <section class="card output-panel">
        <h3>输出</h3>
        <pre id="output">就绪</pre>
      </section>
    </main>
  </div>
</template>
