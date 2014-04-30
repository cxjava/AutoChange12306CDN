function withjQuery(callback, safe) {

	if (typeof(jQuery) == "undefined") {

		var script = document.createElement("script");
		script.type = "text/javascript";
		script.src = "https://lib.sinaapp.com/js/jquery/1.7.2/jquery.min.js";

		if (safe) {
			var cb = document.createElement("script");
			cb.type = "text/javascript";
			cb.textContent = "jQuery.noConflict();(" + callback.toString() + ")(jQuery, window);";

			script.addEventListener('load', function() {

				document.head.appendChild(cb);

			});
		} else {
			var dollar = undefined;
			if (typeof($) != "undefined") dollar = $;
			script.addEventListener('load', function() {
				jQuery.noConflict();
				$ = dollar;
				callback(jQuery, window);
			});
		}
		document.head.appendChild(script);
	} else {

		setTimeout(function() {
			//Firefox supports
			callback(jQuery, typeof unsafeWindow === "undefined" ? window : unsafeWindow);
		}, 30);
	}
}


autoSearchTime = 700;
withjQuery(function($, window) {
	$("div.yzm input#randCode2").unbind('keyup');

	function bO() {
		if ($("#sf2").is(":checked")) {
			return "0X00"
		} else {
			return "ADULT"
		}
	};
	$("div.yzm input#randCode2").on("click",
		function(bR) {
			$.ajax({
				url: ctx + "passcodeNew/checkRandCodeAnsyn",
				type: "post",
				data: {
					randCode: $("div.yzm input#randCode2").val(),
					rand: "sjrand"
				},
				async: false,
				success: function(bS) {
					bb = $("div.yzm input#randCode2").val();
					$("#back_edit").trigger("click");
					$.ajax({
						url: ctx + "confirmPassenger/confirmSingleForQueueAsys",
						type: "post",
						data: {
							passengerTicketStr: getpassengerTicketsForAutoSubmit(),
							oldPassengerStr: getOldPassengersForAutoSubmit(),
							randCode: $("#randCode").val(),
							purpose_codes: bO(),
							key_check_isChange: md5Str,
							leftTicketStr: leftTicketStr,
							train_location: location_code,
							_json_att: ""
						},
						dataType: "json",
						async: true,
						success: function(bR) {
							otsRedirect("post", ctx + "payOrder/init?random=" + new Date().getTime(), {})
						},
						error: function(bR, bT, bS) {
							return
						}
					})
					$("div.yzm input#randCode2").removeClass("inptxt w100 error").addClass("inptxt w100");
					$("#i-ok2").css("display", "block");
					$("#c_error2").html("");
					$("#c_error2").removeClass("error");
					return
				}
			})

			bb = $("div.yzm input#randCode2").val()
		});
	setInterval(function() {

		if ($('#autosubmitcheckticketinfo').css('display') != 'none') {
			$("div.yzm input#randCode2").val('xxoo');
			$("div.yzm input#randCode2").trigger('click');
		}
	}, 400);
}, true);