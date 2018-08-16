<?php
for ($i=0;$i<3;$i++) {
    $pdo = new pdo("mysql:host=10.10.62.28;dbname=showapp", "root", "sd-9898w", array(PDO::ATTR_AUTOCOMMIT => 0));
    $pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
    $row = $pdo->query("select * from banner where 1 limit 1", PDO::A);
    $data = $row->fetch();
}