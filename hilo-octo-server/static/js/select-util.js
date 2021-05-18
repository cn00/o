(function(global) {
    'use strict';

    function getSelectedFiles() {
        var ids = [],
            filenames = [];

        $('tr.is-selected').each(function() {
            ids.push($(this).attr('data-octo-fileId'));
            filenames.push($(this).attr('data-octo-filename'));
        });

        return {
            ids: ids,
            filenames: filenames
        };
    }

    global.getSelectedFiles = getSelectedFiles;
}(this));
