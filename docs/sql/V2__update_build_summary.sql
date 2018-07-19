alter table build_summary add column status int;
update build_summary as bs
  set status = (case
                  when bs.failed = false
                  then 4
                  else 3
                end)
;