import { useMemo } from 'react';
import './Heatmap.css';

const MONTHS = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];

function getIntensity(count) {
  if (count === 0) return 0;
  if (count <= 1) return 1;
  if (count <= 3) return 2;
  if (count <= 5) return 3;
  return 4;
}

export default function Heatmap({ data = {} }) {
  const { weeks, monthLabels, totalSubmissions } = useMemo(() => {
    const today = new Date();
    // Start at exactly midnight local time to avoid timezone offset shifts
    today.setHours(0, 0, 0, 0);
    
    const startDate = new Date(today);
    startDate.setDate(startDate.getDate() - 364);

    // Adjust start to Sunday
    const dayOfWeek = startDate.getDay();
    startDate.setDate(startDate.getDate() - dayOfWeek);

    const weeks = [];
    let currentDate = new Date(startDate);
    const monthLabels = [];
    let lastMonth = -1;
    let total = 0;

    // We build precisely 53 weeks (or 52) up to today's week
    while (currentDate <= today || currentDate.getDay() !== 0) {
      const week = [];

      for (let d = 0; d < 7; d++) {
        // Correct Local Date String (YYYY-MM-DD) instead of UTC toISOString()
        const yyyy = currentDate.getFullYear();
        const mm = String(currentDate.getMonth() + 1).padStart(2, '0');
        const dd = String(currentDate.getDate()).padStart(2, '0');
        const dateStr = `${yyyy}-${mm}-${dd}`;
        
        const count = data[dateStr] || 0;
        total += count;

        // If month changes, record it at this week's index
        if (currentDate.getMonth() !== lastMonth) {
          // Avoid duplicate labels for the same week (e.g. if Jan 1 is late in the week)
          if (monthLabels.length === 0 || monthLabels[monthLabels.length - 1].weekIndex !== weeks.length) {
            monthLabels.push({
              label: MONTHS[currentDate.getMonth()],
              weekIndex: weeks.length,
            });
          }
          lastMonth = currentDate.getMonth();
        }

        week.push({
          date: dateStr,
          count,
          intensity: getIntensity(count),
          isInRange: currentDate <= today,
        });

        currentDate.setDate(currentDate.getDate() + 1);
      }

      weeks.push(week);
      if (currentDate > today) break;
    }

    // Filter out labels that are clumped together (e.g. at the start of the year)
    const spacedLabels = monthLabels.filter((m, i, arr) => 
      i === arr.length - 1 || (arr[i + 1].weekIndex - m.weekIndex) >= 3
    );

    return { weeks, monthLabels: spacedLabels, totalSubmissions: total };
  }, [data]);

  return (
    <div className="heatmap-container" id="submission-heatmap">
      <div className="heatmap-header">
        <span className="heatmap-total">
          <strong>{totalSubmissions}</strong> submissions in the last year
        </span>
      </div>

      <div className="heatmap-scroll">
        <div className="heatmap-months">
          {monthLabels.map((m, i) => (
            <span
              key={i}
              className="month-label"
              style={{ gridColumnStart: m.weekIndex + 1 }}
            >
              {m.label}
            </span>
          ))}
        </div>

        <div className="heatmap-grid">
          <div className="weekday-labels">
            <span></span>
            <span>Mon</span>
            <span></span>
            <span>Wed</span>
            <span></span>
            <span>Fri</span>
            <span></span>
          </div>

          <div className="heatmap-weeks">
            {weeks.map((week, wi) => (
              <div key={wi} className="heatmap-week">
                {week.map((day, di) => (
                  <div
                    key={di}
                    className={`heatmap-day intensity-${day.isInRange ? day.intensity : 'empty'}`}
                    title={day.isInRange ? `${day.count} submissions on ${day.date}` : ''}
                  />
                ))}
              </div>
            ))}
          </div>
        </div>
      </div>

      <div className="heatmap-legend">
        <span className="legend-label">Less</span>
        {[0, 1, 2, 3, 4].map((i) => (
          <div key={i} className={`heatmap-day intensity-${i}`} />
        ))}
        <span className="legend-label">More</span>
      </div>
    </div>
  );
}
