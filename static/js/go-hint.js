(function(mod) {
	mod(CodeMirror);
})(function(CodeMirror) {
  	"use strict";

  	CodeMirror.registerHelper("hint", "go", function(editor, options) {
    	var cur = editor.getCursor();
	
		var data = {
			code: editor.getValue(),
			cursorLine: editor.getCursor().line,
			cursorCh: editor.getCursor().ch
		};
					
		editorWS.send(JSON.stringify(data));
		
		return {list: autocompleteHints, from: cur, to: cur};
  	});
});
