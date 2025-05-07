package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"github.com/telekom/controlplane-mono/common-server/pkg/problems"
)

func ReturnWithProblem(ctx *fiber.Ctx, problem problems.Problem, err error) error {
	if problem != nil {
		return ctx.Status(problem.Code()).JSON(problem, "application/problem+json")
	}

	return ReturnWithError(ctx, err)
}

func ReturnWithError(c *fiber.Ctx, err error) error {
	var p problems.Problem
	if errors.As(err, &p) {
		return c.Status(p.Code()).JSON(p, "application/problem+json")
	}
	var fe *fiber.Error
	if errors.As(err, &fe) {
		if fe.Code >= 500 {
			p = problems.NewProblemOfError(err)
		} else if fe.Code == 404 {
			p = problems.NotFound()
		} else if fe.Code == 405 {
			p = problems.MethodNotAllowed(c.Method())
		} else {
			p = problems.BadRequest(fe.Message)
		}
		return c.Status(fe.Code).JSON(p, "application/problem+json")
	}

	problem := problems.NewProblemOfError(err)
	return c.Status(problem.Code()).JSON(problem, "application/problem+json")
}
