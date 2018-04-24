package log

const (
  TEMPLATE_LOAD		= "package=%s, method=%s, message[%v]"
  TEMPLATE_REQUEST	= "transaction_id=%+v, package=%s, struct[%s], path=%s, method=%s, method_args[%v]"
  TEMPLATE_REQUEST_MSG	= "transaction_id=%+v, package=%s, struct[%s], path=%s, method=%s, method_args[%v], message[%+v]"
  TEMPLATE_RESPONSE	= "transaction_id=%+v, package=%s, struct[%s], path=%s, method=%s, method_resp=%s status_code=%d, message[%+v]"
  TEMPLATE_INTEGRATION	= "transaction_id=%+v, package=%s, struct[%s], integration=%s, integration_method=%s, integration_args[%+v]"
  TEMPLATE_PUBLISH	= "transaction_id=%+v, package=%s, method=%s, integration=%s, exchange=%s, routing=%s, message[%+v]"
  TEMPLATE_CORE		= "transaction_id=%+v, package=%s, method=%s, integration=%s, message[%+v]"
  TEMPLATE_LOG_CORE	= "transaction_id=%+v, package=%s, method=%s, integration=%s, type=%s, integration_args[%+v], message[%+v]"
  EMPTY_STR		= ""
  INIT			= "Init"
  DONE			= "Done"
)
