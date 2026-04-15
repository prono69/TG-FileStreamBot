package commands

import (
	"EverythingSuckz/fsb/config"
	"EverythingSuckz/fsb/internal/utils"
	"fmt"
	"math"
	"net/http"
	"runtime"
	"time"

	"github.com/celestix/gotgproto/dispatcher"
	"github.com/celestix/gotgproto/dispatcher/handlers"
	"github.com/celestix/gotgproto/ext"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

var startTime = time.Now()

func (m *command) LoadUsage(dispatcher dispatcher.Dispatcher) {
	log := m.log.Named("usage")
	defer log.Sugar().Info("Loaded")
	dispatcher.AddHandler(handlers.NewCommand("usage", usage))
}

// ---------- utils ----------

func formatBytes(bytes uint64) string {
	if bytes == 0 {
		return "0 B"
	}
	units := []string{"B", "KB", "MB", "GB", "TB"}
	i := int(math.Log(float64(bytes)) / math.Log(1024))
	if i >= len(units) {
		i = len(units) - 1
	}
	val := float64(bytes) / math.Pow(1024, float64(i))
	return fmt.Sprintf("%.2f %s", val, units[i])
}

func formatUptime(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, seconds)
	}
	return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
}

// ---------- command ----------

func usage(ctx *ext.Context, u *ext.Update) error {
	defer func() {
		if r := recover(); r != nil {
			ctx.Reply(u, ext.ReplyTextString("⚠️ Failed to fetch usage stats"), nil)
		}
	}()

	chatId := u.EffectiveChat().GetID()

	// ✅ Safe private check (no PeerStorage bug)
	if u.EffectiveChat().GetType() != "private" {
		return dispatcher.EndGroups
	}

	// ✅ Allowed users
	if len(config.ValueOf.AllowedUsers) != 0 && !utils.Contains(config.ValueOf.AllowedUsers, chatId) {
		ctx.Reply(u, ext.ReplyTextString("You are not allowed to use this bot."), nil)
		return dispatcher.EndGroups
	}

	// ---------- stats ----------

	cpuPercent, _ := cpu.Percent(0, false)
	cpuCount, _ := cpu.Counts(true)
	cpuUsage := 0.0
	if len(cpuPercent) > 0 {
		cpuUsage = cpuPercent[0]
	}

	memStats, _ := mem.VirtualMemory()
	diskStats, _ := disk.Usage("/")

	netStats, _ := net.IOCounters(false)
	var bytesSent, bytesRecv uint64
	if len(netStats) > 0 {
		bytesSent = netStats[0].BytesSent
		bytesRecv = netStats[0].BytesRecv
	}

	var rtm runtime.MemStats
	runtime.ReadMemStats(&rtm)

	uptime := time.Since(startTime)
	goroutines := runtime.NumGoroutine()

	// ---------- server check ----------

	serverStatus := "🔴 Offline"
	client := http.Client{Timeout: 3 * time.Second}

	resp, err := client.Get("https://ddl.ichigo.eu.org")
	if err == nil && resp.StatusCode < 500 {
		serverStatus = "🟢 Online"
	} else if err != nil {
		serverStatus = "🟡 Unreachable"
	}

	// ---------- message ----------

	msg := fmt.Sprintf(
		"📊 **FSB Usage Stats**\n"+
			"━━━━━━━━━━━━━━━━━━\n\n"+

			"⏱ **Uptime**\n"+
			"└ `%s`\n\n"+

			"🖥 **CPU**\n"+
			"├ Cores: `%d`\n"+
			"└ Usage: `%.1f%%`\n\n"+

			"🧠 **Memory**\n"+
			"├ Used:  `%s / %s`\n"+
			"└ Usage: `%.1f%%`\n\n"+

			"💾 **Disk**\n"+
			"├ Used:  `%s / %s`\n"+
			"└ Usage: `%.1f%%`\n\n"+

			"🌐 **Network**\n"+
			"├ Upload:   `%s`\n"+
			"└ Download: `%s`\n\n"+

			"⚙️ **Runtime**\n"+
			"├ Go Version:  `%s`\n"+
			"├ Goroutines:  `%d`\n"+
			"└ Heap Alloc:  `%s`\n\n"+

			"🤖 **Server**\n"+
			"└ Status: %s\n"+
			"━━━━━━━━━━━━━━━━━━",

		formatUptime(uptime),
		cpuCount, cpuUsage,
		formatBytes(memStats.Used), formatBytes(memStats.Total), memStats.UsedPercent,
		formatBytes(diskStats.Used), formatBytes(diskStats.Total), diskStats.UsedPercent,
		formatBytes(bytesSent), formatBytes(bytesRecv),
		runtime.Version(), goroutines, formatBytes(rtm.HeapAlloc),
		serverStatus,
	)

	ctx.Reply(u, ext.ReplyTextString(msg), nil)
	return dispatcher.EndGroups
}