resource "swo_compositemetric" "disk_io_rate" {
  name         = "composite.custom.system.disk.io.rate"
  display_name = "Disk IO rate"
  description  = "Disk bytes transferred per second"
  formula      = "rate(system.disk.io[5m])"
  unit         = "bytes/s"
}