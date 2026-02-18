package launcher

const jxaScreenDetect = `
ObjC.import("AppKit");
var term = Application("Terminal");
var b = term.windows[0].bounds();
var winX = b.x, winY = b.y;
var screens = $.NSScreen.screens;
var mainH = screens.objectAtIndex(0).frame.size.height;
var result = "";
for (var i = 0; i < screens.count; i++) {
    var f = screens.objectAtIndex(i).frame;
    var sx = f.origin.x;
    var sy = mainH - f.origin.y - f.size.height;
    var sw = f.size.width;
    var sh = f.size.height;
    if (winX >= sx && winX < sx + sw && winY >= sy && winY < sy + sh) {
        result = Math.round(sx) + " " + Math.round(sy) + " " + Math.round(sx + sw) + " " + Math.round(sy + sh);
        break;
    }
}
result`

const tilingScriptTemplate = `tell application "Terminal"
    activate
end tell

delay 0.5

set screenX to {{.X1}}
set screenY to {{.Y1}}
set screenWidth to {{.X2}}
set screenHeight to {{.Y2}}
set menuBar to 25
set rowColsList to { {{.RowCols}} }
set numRows to {{.NumRows}}
set cellH to (screenHeight - screenY - menuBar) / numRows

tell application "Terminal"
    repeat with r from 1 to numRows
        set thisCols to item r of rowColsList
        set cellW to (screenWidth - screenX) / thisCols
        repeat with c from 0 to (thisCols - 1)
            set x1 to (screenX + c * cellW) as integer
            set x2 to (screenX + (c + 1) * cellW) as integer
            set y1 to (screenY + menuBar + (r - 1) * cellH) as integer
            set y2 to (screenY + menuBar + r * cellH) as integer

            do script "{{.TermCmd}}"
            delay 0.3
            set bounds of window 1 to {x1, y1, x2, y2}
        end repeat
    end repeat
end tell`
