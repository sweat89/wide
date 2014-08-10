(function(mod) {
	mod(CodeMirror);
})(function(CodeMirror) {
  	"use strict";

  	CodeMirror.registerHelper("hint", "go", function(editor, options) {
    	var cur = editor.getCursor();
	
		var request = {
			code: editor.getValue(),
			
			cursorLine: editor.getCursor().line,
			cursorCh: editor.getCursor().ch
		};
		
		// XXX: 回调有问题，暂时不使用 WS 协议
		//editorWS.send(JSON.stringify(request));
		
		var autocompleteHints = [];
		
		$.ajax({
			async: false, // 同步执行
			type: 'POST',
			url: '/autocomplete',
			data: JSON.stringify(request),
			dataType: "json",
			success: function(data) {
				var autocompleteArray = data[1];
					
				for	(var i = 0; i < autocompleteArray.length; i++) {
					autocompleteHints[i] = autocompleteArray[i].name;
				}
			}
		});
		
		return {list: autocompleteHints, from: cur, to: cur};
  	});
});
