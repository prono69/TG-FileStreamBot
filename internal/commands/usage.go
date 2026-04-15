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

// ---------- LOADER ----------

func (m *command) LoadUsage(dispatcher dispatcher.Dispatcher) {
	log := m.log.Named("usage")
	defer log.Sugar().Info("Loaded")
	dispatcher.AddHandler(handlers.NewCommand("usage", usage))
}

// ---------- UTILS ----------

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

// ---------- COMMAND ----------

func usage(ctx *ext.Context, u *ext.Update) error {
	defer func() {
		if r := recover(); r != nil {
			ctx.Reply(u, ext.ReplyTextString("⚠️ Failed to fetch usage stats"), nil)
		}
	}()

	chatId := u.EffectiveChat().GetID()

	// Allowed users (optional restriction)
	if len(config.ValueOf.AllowedUsers) != 0 && !utils.Contains(config.ValueOf.AllowedUsers, chatId) {
		ctx.Reply(u, ext.ReplyTextString("You are not allowed to use this bot."), nil)
		return dispatcher.EndGroups
	}

	// ---------- SAFE STATS ----------

	// CPU
	cpuUsage := 0.0
	cpuCount := 0

	if c, err := cpu.Counts(true); err == nil {
		cpuCount = c
	}
	if p, err := cpu.Percent(0, false); err == nil && len(p) > 0 {
		cpuUsage = p[0]
	}

	// Memory
	var memUsed, memTotal uint64
	var memPercent float64

	if m, err := mem.VirtualMemory(); err == nil {
		memUsed = m.Used
		memTotal = m.Total
		memPercent = m.UsedPercent
	}

	// Disk
	var diskUsed, diskTotal uint64
	var diskPercent float64

	if d, err := disk.Usage("/"); err == nil {
		diskUsed = d.Used
		diskTotal = d.Total
		diskPercent = d.UsedPercent
	}

	// Network
	var bytesSent, bytesRecv uint64

	if n, err := net.IOCounters(false); err == nil && len(n) > 0 {
		bytesSent = n[0].BytesSent
		bytesRecv = n[0].BytesRecv
	}

	// Runtime
	var rtm runtime.MemStats
	runtime.ReadMemStats(&rtm)

	uptime := time.Since(startTime)
	goroutines := runtime.NumGoroutine()

	// ---------- SERVER CHECK ----------

	serverStatus := "🔴 Offline"
	client := http.Client{Timeout: 3 * time.Second}

	if resp, err := client.Get("https://ddl.ichigo.eu.org"); err == nil {
		if resp.StatusCode < 500 {
			serverStatus = "🟢 Online"
		}
	} else {
		serverStatus = "🟡 Unreachable"
	}

	// ---------- MESSAGE ----------

	msg := fmt.Sprintf(
		"📊 **FSB Usage Stats**\n"+
			"━━━━━━━━━━━━━━━━━━\n\n"+

			"⏱ **Uptime**\n"+
			"└ `%s`\n\n"+

			"🖥 **CPU**\n"+
			"├ Cores: `%d`\n"+
			"└ Usage: `%.1f%%`\n\n"+

			"🧠 **Memory**\n"+
			"├ `%s / %s`\n"+
			"└ `%.1f%%`\n\n"+

			"💾 **Disk**\n"+
			"├ `%s / %s`\n"+
			"└ `%.1f%%`\n\n"+

			"🌐 **Network**\n"+
			"├ Upload: `%s`\n"+
			"└ Download: `%s`\n\n"+

			"⚙️ **Runtime**\n"+
			"├ `%s`\n"+
			"├ Goroutines: `%d`\n"+
			"└ Heap: `%s`\n\n"+

			"🤖 **Server**\n"+
			"└ %s\n"+

			"━━━━━━━━━━━━━━━━━━",

		formatUptime(uptime),
		cpuCount, cpuUsage,
		formatBytes(memUsed), formatBytes(memTotal), memPercent,
		formatBytes(diskUsed), formatBytes(diskTotal), diskPercent,
		formatBytes(bytesSent), formatBytes(bytesRecv),
		runtime.Version(), goroutines, formatBytes(rtm.HeapAlloc),
		serverStatus,
	)

	ctx.Reply(u, ext.ReplyTextString(msg), nil)
	return dispatcher.EndGroups
}