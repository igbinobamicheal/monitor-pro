use super::collector::MetricSnapshot;
use std::collections::VecDeque;

pub struct TelemetryBuffer {
    capacity: usize,
    queue: VecDeque<MetricSnapshot>,
}

impl TelemetryBuffer {
    pub fn new(capacity: usize) -> Self {
        Self {
            capacity,
            queue: VecDeque::with_capacity(capacity),
        }
    }

    pub fn push(&mut self, snapshot: MetricSnapshot) {
        if self.queue.len() >= self.capacity {
            self.queue.pop_front(); // FIFO eviction when buffer is saturated
        }
        self.queue.push_back(snapshot);
    }

    pub fn drain_all(&mut self) -> Vec<MetricSnapshot> {
        self.queue.drain(..).collect()
    }

    pub fn len(&self) -> usize {
        self.queue.len()
    }

    pub fn is_empty(&self) -> bool {
        self.queue.is_empty()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_ring_buffer_fifo_eviction() {
        let mut buf = TelemetryBuffer::new(3);
        for i in 1..=5 {
            buf.push(MetricSnapshot {
                cpu_usage: i as f64 * 10.0,
                ram_usage: 50.0,
                disk_usage: 50.0,
                network_in_kb: 0.0,
                network_out_kb: 0.0,
                uptime_seconds: 100,
                timestamp: i,
            });
        }

        assert_eq!(buf.len(), 3);
        let items = buf.drain_all();
        assert_eq!(items.len(), 3);
        assert_eq!(items[0].timestamp, 3);
        assert_eq!(items[1].timestamp, 4);
        assert_eq!(items[2].timestamp, 5);
        assert!(buf.is_empty());
    }
}
