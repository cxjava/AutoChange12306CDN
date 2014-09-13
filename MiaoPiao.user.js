/*
 *  12306 Auto Query => A javascript snippet to help you book tickets online.
 *  12306 Booking Assistant
 *  Copyright (C) 2011 Hidden
 *
 *  12306 Auto Query => A javascript snippet to help you book tickets online.
 *  Copyright (C) 2011 Jingqin Lynn
 *
 *  12306 Auto Login => A javascript snippet to help you auto login 12306.com.
 *  Copyright (C) 2011 Kevintop
 *
 *  Includes jQuery
 *  Copyright 2011, John Resig
 *  Dual licensed under the MIT or GPL Version 2 licenses.
 *  http://jquery.org/license
 *
 *  This program is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU General Public License as published by
 *  the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 *  This program is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *  GNU General Public License for more details.
 *
 *  You should have received a copy of the GNU General Public License
 *  along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

// ==UserScript==
// @name         12306 果果秒票助手 V3
// @version		 3.0
// @author       guozili@163.com
// @namespace    http://www.guozili.25u.com
// @description  帮您秒票的小助手 by guozili@163.com
// @include      *://kyfw.12306.cn/otn/*
// @require	https://ajax.aspnetcdn.com/ajax/jquery/jquery-1.7.1.min.js
// ==/UserScript==

function withjQuery(callback, safe) {

	if (typeof(jQuery) == "undefined") {

		var script = document.createElement("script");
		script.type = "text/javascript";
		script.src = "https://ajax.aspnetcdn.com/ajax/jquery/jquery-1.7.1.min.js";

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

withjQuery(function($, window) {

	function route(match, fn) {
		if (window.location.href.indexOf(match) != -1) {
			fn();
		}
	}

	var clickevent = document.createEvent('MouseEvents');
	clickevent.initEvent('click', false, true);

	function query() {

		setTimeout(function() {
			window.autoSearchTime = 1000;
		}, 2000);

		setInterval(function() {
			if ($("#autosubmitcheckticketinfo").css("display") == "none" && $("#query_ticket").text() == "停止查询") {
				//停止
				document.getElementById("query_ticket").dispatchEvent(clickevent);
				//继续查询
				setTimeout(function() {
					document.getElementById("query_ticket").dispatchEvent(clickevent);
				}, 500);
			}
		}, 4000);

	}

	route("init", query);

}, true);