# Adaptive Learning Setup Guide

## Quick Start (Optional Persistence)

### 1. Create Input Number Helper

In Home Assistant:
1. Go to **Settings → Devices & Services → Helpers**
2. Click **"+ Create Helper" → Number**
3. Configure:
   - **Name**: `Adaptive Comfort Learned Offset`
   - **Minimum**: `-10`
   - **Maximum**: `10`
   - **Step**: `0.01`
   - **Unit**: `°F` (or `°C` depending on your system)
   - **Icon**: `mdi:brain` (optional)

### 2. Configure Blueprint

In your automation:
1. **Manual Override Detection** section:
   - ✅ **Enable Manual Override Detection**: On
   - **Override Duration**: 60 minutes (adjust as needed)
   - ✅ **Learn from Manual Adjustments**: On
   - **Learning Rate**: 0.15 (start conservative)
   - **Learned Offset Storage**: Select `input_number.adaptive_comfort_learned_offset`

## How It Works

### The Learning Process

```
User adjusts thermostat: 70°F → 73°F
Blueprint predicted: 70°F
Error: +3°F

Calculation:
new_offset = 0.85 * old_offset + 0.15 * (+3°F)
           = 0.85 * 0 + 0.45
           = +0.45°F

Future predictions: now +0.45°F warmer
```

### Convergence Timeline

With **α = 0.15** (default):
- After 1 adjustment: ~15% adapted
- After 5 adjustments: ~56% adapted  
- After 10 adjustments: ~80% adapted
- After 15 adjustments: ~93% adapted

### Learning Rate Guide

| Rate | Speed | Stability | Best For |
|------|-------|-----------|----------|
| 0.05 | Very Slow | Very Stable | Highly consistent preferences |
| 0.10 | Slow | Stable | Most users |
| **0.15** | **Moderate** | **Balanced** | **Recommended default** |
| 0.20 | Fast | Moderate | Quickly adapting to new patterns |
| 0.30 | Very Fast | Less Stable | Experimental/testing |

## Advanced Configuration

### For Colorado (Mixed-Dry Climate)

Your regional presets are already set:
- **Winter Bias**: +0.3°C (keeps you warmer in cold months)
- **Summer Bias**: -0.2°C (keeps you cooler in hot months)
- **Shoulder Bias**: 0°C (neutral in spring/fall)

The learned offset **adds to** these biases, so:
```
Final Comfort = ASHRAE-55 + Regional Bias + Learned Offset + Sleep + CO₂
```

### Without Persistence

If you don't create the helper:
- Learning still works during the current session
- Offset resets to 0 when Home Assistant restarts
- Useful for testing before committing to persistence

## Monitoring Learning

### Debug Logs

Enable **Debug Level: basic** or **verbose** to see:

```
Manual override detected (climate_manual_change). 
Manual=21.7°C, Predicted=20.0°C, Error=+1.7°C. 
Learned offset: +0.5°C → +0.76°C.
Pausing for 60 min.
```

### Dashboard Card (Optional)

Add to your dashboard to visualize:

```yaml
type: entities
entities:
  - entity: input_number.adaptive_comfort_learned_offset
    name: Learned Temperature Offset
```

## Tuning Tips

### Too Aggressive?
- Reduce learning rate to 0.10 or 0.05
- Increases stability, slower adaptation

### Too Slow?
- Increase learning rate to 0.20 or 0.25
- Faster adaptation, slightly less stable

### Reset Learning
Set the helper to `0` to start fresh.

## FAQ

**Q: Does this replace seasonal adaptation?**  
No! Learned offset adds to your Colorado regional biases. Both work together.

**Q: What if I make random adjustments?**  
The exponential average smooths out noise. Consistent patterns emerge over time.

**Q: Can I disable learning temporarily?**  
Yes, just toggle "Learn from Manual Adjustments" to off in the blueprint config.

**Q: Does it learn time-of-day preferences?**  
Not yet. Current implementation learns a single global offset. Future versions could add contextual learning (morning vs evening, etc.).

**Q: What happens during sleep mode?**  
Sleep bias is separate and still applied. Learning tracks your general preference across all modes.

## Troubleshooting

**Learning not working:**
1. Check "Learn from Manual Adjustments" is enabled
2. Verify helper entity is selected and exists
3. Ensure manual changes exceed tolerance (default: 1.0°)
4. Check debug logs for offset updates

**Helper value not changing:**
- Check helper minimum/maximum range (-10 to +10 recommended)
- Verify no conflicting automations writing to the same helper

**Offset seems wrong:**
- Reset helper to 0 and let it re-learn
- Reduce learning rate for more conservative updates
- Check debug logs to see error calculations
